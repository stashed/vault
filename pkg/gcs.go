/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"

	kmsv1 "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	ServiceAccountJSON    = "sa.json"
	GoogleApplicationCred = "GOOGLE_APPLICATION_CREDENTIALS"
)

func (opt *VaultOptions) newGcsClient(cred string) (*storage.Client, error) {
	secret, err := opt.kubeClient.CoreV1().Secrets(opt.appBindingNamespace).Get(context.TODO(), cred, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if _, ok := secret.Data[ServiceAccountJSON]; !ok {
		return nil, errors.Errorf("%s not found in secret", ServiceAccountJSON)
	}

	path := filepath.Join("/tmp", fmt.Sprintf("google-sa-cred-%s", randomString(6)))
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}

	saFile := filepath.Join(path, ServiceAccountJSON)
	if err = os.WriteFile(saFile, secret.Data[ServiceAccountJSON], os.ModePerm); err != nil {
		return nil, err
	}

	if err = os.Setenv(GoogleApplicationCred, saFile); err != nil {
		return nil, err
	}

	client, err := storage.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (opt *VaultOptions) getGcsTokenKeys() (map[string]string, error) {
	client, err := opt.newGcsClient(opt.credentialSecretRef)
	if err != nil {
		return nil, err
	}

	keys := opt.getKeys()
	for key := range keys {
		rc, err := client.Bucket(opt.bucket).Object(key).NewReader(context.TODO())
		if err != nil {
			return nil, err
		}
		rc.Close()

		body, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
			opt.kmsProject, opt.kmsLocation,
			opt.kmsKeyRing, opt.kmsCryptoKey)

		decryptedToken, err := decryptSymmetric(name, body)
		if err != nil {
			return nil, err
		}

		keys[key] = decryptedToken
	}

	return keys, nil
}

func (opt *VaultOptions) setGcsTokenKeys(vs *vaultapi.VaultServer, keys map[string]string) error {
	mode := vs.Spec.Unsealer.Mode

	var credRef string
	if mode.GoogleKmsGcs.CredentialSecretRef != nil {
		credRef = mode.GoogleKmsGcs.CredentialSecretRef.Name
	}

	client, err := opt.newGcsClient(credRef)
	if err != nil {
		return err
	}

	kmsService, err := cloudkms.NewService(context.TODO(), option.WithScopes(cloudkms.CloudPlatformScope))
	if err != nil {
		return errors.Errorf("error creating google kms service client: %s", err.Error())
	}

	name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		mode.GoogleKmsGcs.KmsProject, mode.GoogleKmsGcs.KmsLocation,
		mode.GoogleKmsGcs.KmsKeyRing, mode.GoogleKmsGcs.KmsCryptoKey)

	for key, value := range keys {
		resp, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.Encrypt(name, &cloudkms.EncryptRequest{
			Plaintext: base64.StdEncoding.EncodeToString([]byte(value)),
		}).Do()
		if err != nil {
			return errors.Errorf("error encrypting data: %s", err.Error())
		}

		cipherText, err := base64.StdEncoding.DecodeString(resp.Ciphertext)
		if err != nil {
			return err
		}

		w := client.Bucket(mode.GoogleKmsGcs.Bucket).Object(key).NewWriter(context.TODO())
		if _, err := w.Write(cipherText); err != nil {
			return fmt.Errorf("error writing key '%s' to gcs bucket '%s'", key, mode.GoogleKmsGcs.Bucket)
		}

		if err = w.Close(); err != nil {
			return err
		}
	}

	return nil
}

func decryptSymmetric(name string, ciphertext []byte) (string, error) {
	client, err := kmsv1.NewKeyManagementClient(context.TODO())
	if err != nil {
		return "", errors.Errorf("failed to create kms client: %v", err)
	}
	defer client.Close()

	crc32c := func(data []byte) uint32 {
		t := crc32.MakeTable(crc32.Castagnoli)
		return crc32.Checksum(data, t)
	}
	ciphertextCRC32C := crc32c(ciphertext)

	req := &kmspb.DecryptRequest{
		Name:             name,
		Ciphertext:       ciphertext,
		CiphertextCrc32C: wrapperspb.Int64(int64(ciphertextCRC32C)),
	}

	result, err := client.Decrypt(context.TODO(), req)
	if err != nil {
		return "", errors.Errorf("failed to decrypt ciphertext with %s", err.Error())
	}

	if int64(crc32c(result.Plaintext)) != result.PlaintextCrc32C.Value {
		return "", errors.Errorf("decrypt response corrupted in-transit")
	}

	return string(result.Plaintext), nil
}
