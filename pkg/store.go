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
	"math/rand"
	"os"
	"path/filepath"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	UnsealModeGoogleKmsGcs     = "googleKmsGcs"
	UnsealModeKubernetesSecret = "kubernetesSecret"
)

const (
	ServiceAccountJSON    = "sa.json"
	GoogleApplicationCred = "GOOGLE_APPLICATION_CREDENTIALS"
)

func (opt *VaultOptions) GetTokenKeys() (map[string]string, error) {
	switch opt.OldUnsealMode {
	case UnsealModeGoogleKmsGcs:
		return opt.getGcsTokenKeys()
	case UnsealModeKubernetesSecret:
		return opt.getK8sTokenKeys()
	}

	return nil, nil
}

func (opt *VaultOptions) SetTokenKeys(keys map[string]string) error {
	switch opt.NewUnsealMode {
	case UnsealModeGoogleKmsGcs:
		return opt.setGcsTokenKeys(keys)
	case UnsealModeKubernetesSecret:
		return opt.setK8sTokenKeys(keys)
	}

	return errors.New("unknown unseal mode")
}

func (opt *VaultOptions) getGcsTokenKeys() (map[string]string, error) {
	secret, err := opt.KubeClient.CoreV1().Secrets(opt.AppBindingNamespace).Get(context.TODO(), opt.OldCredentialSecretRef, metav1.GetOptions{})
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

	keys := make(map[string]string)

	var key string
	key = opt.tokenName()
	keys[key] = ""
	for id := 0; int64(id) < opt.SecretShares; id++ {
		key = opt.unsealKeyName(id)
		keys[key] = ""
	}

	klog.Infoln("keys; ", keys)
	for key = range keys {
		rc, err := client.Bucket(opt.OldBucket).Object(key).NewReader(context.TODO())
		if err != nil {
			return nil, err
		}
		rc.Close()

		body, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
			opt.OldKmsProject, opt.OldKmsLocation,
			opt.OldKmsKeyRing, opt.OldKmsCryptoKey)

		decryptedToken, err := decryptSymmetric(name, body)
		if err != nil {
			return nil, err
		}

		keys[key] = decryptedToken
	}

	return keys, nil
}

func (opt *VaultOptions) setGcsTokenKeys(keys map[string]string) error {
	secret, err := opt.KubeClient.CoreV1().Secrets(opt.AppBindingNamespace).Get(context.TODO(), opt.NewCredentialSecretRef, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if _, ok := secret.Data[ServiceAccountJSON]; !ok {
		return errors.Errorf("%s not found in secret", ServiceAccountJSON)
	}

	path := filepath.Join("/tmp", fmt.Sprintf("google-sa-cred-%s", randomString(6)))
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	saFile := filepath.Join(path, ServiceAccountJSON)
	if err = os.WriteFile(saFile, secret.Data[ServiceAccountJSON], os.ModePerm); err != nil {
		return err
	}

	if err = os.Setenv(GoogleApplicationCred, saFile); err != nil {
		return err
	}

	client, err := storage.NewClient(context.TODO())
	if err != nil {
		return err
	}

	kmsService, err := cloudkms.NewService(context.TODO(), option.WithScopes(cloudkms.CloudPlatformScope))
	if err != nil {
		return errors.Errorf("error creating google kms service client: %s", err.Error())
	}

	name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		opt.NewKmsProject, opt.NewKmsLocation,
		opt.NewKmsKeyRing, opt.NewKmsCryptoKey)

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

		w := client.Bucket(opt.NewBucket).Object(key).NewWriter(context.TODO())
		if _, err := w.Write(cipherText); err != nil {
			return fmt.Errorf("error writing key '%s' to gcs bucket '%s'", key, opt.NewBucket)
		}

		if err = w.Close(); err != nil {
			return err
		}
	}

	return nil
}

func decryptSymmetric(name string, ciphertext []byte) (string, error) {
	client, err := kms.NewKeyManagementClient(context.TODO())
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

func (opt *VaultOptions) getK8sTokenKeys() (map[string]string, error) {
	secret, err := opt.KubeClient.CoreV1().Secrets(opt.AppBindingNamespace).Get(context.TODO(), opt.OldSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	keys := make(map[string]string)

	var key string
	key = opt.tokenName()
	keys[key] = ""
	for id := 0; int64(id) < opt.SecretShares; id++ {
		key = opt.unsealKeyName(id)
		keys[key] = ""
	}

	for k, v := range secret.Data {
		keys[k] = string(v)
	}

	return keys, nil
}

func (opt *VaultOptions) setK8sTokenKeys(keys map[string]string) error {
	secret, err := opt.KubeClient.CoreV1().Secrets(opt.AppBindingNamespace).Get(context.TODO(), opt.NewSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for key, value := range keys {
		secret.Data[key] = []byte(value)
	}

	_, err = opt.KubeClient.CoreV1().Secrets(opt.AppBindingNamespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	return err
}

func (opt *VaultOptions) unsealKeyName(id int) string {
	return fmt.Sprintf("%s-unseal-key-%d", opt.KeyPrefix, id)
}

func (opt *VaultOptions) tokenName() string {
	return fmt.Sprintf("%s-root-token", opt.KeyPrefix)
}

func randomString(n int) string {
	rand.Seed(time.Now().Unix())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
