/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package google_kms_gcs

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stash.appscode.dev/vault/pkg"
	"stash.appscode.dev/vault/pkg/token-keys-store/api"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	ServiceAccountJSON    = "sa.json"
	GoogleApplicationCred = "GOOGLE_APPLICATION_CREDENTIALS"
)

type TokenKeyInfo struct {
	storageClient *storage.Client
	kubeClient    kubernetes.Interface
	opt           *pkg.VaultOptions
}

var _ api.TokenKeyInterface = &TokenKeyInfo{}

func Old(opt *pkg.VaultOptions) (*TokenKeyInfo, error) {
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

	return &TokenKeyInfo{
		storageClient: client,
		opt:           opt,
	}, nil
}

func New(opt *pkg.VaultOptions) (*TokenKeyInfo, error) {
	secret, err := opt.KubeClient.CoreV1().Secrets(opt.AppBindingNamespace).Get(context.TODO(), opt.NewCredentialSecretRef, metav1.GetOptions{})
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

	return &TokenKeyInfo{
		storageClient: client,
		opt:           opt,
	}, nil
}

// get using the old info

func (ti *TokenKeyInfo) Get(key string) (string, error) {
	rc, err := ti.storageClient.Bucket(ti.opt.OldBucket).Object(key).NewReader(context.TODO())
	if err != nil {
		return "", err
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		ti.opt.OldKmsProject, ti.opt.OldKmsLocation,
		ti.opt.OldKmsKeyRing, ti.opt.OldKmsCryptoKey)

	decryptedToken, err := decryptSymmetric(name, body)
	if err != nil {
		return "", err
	}

	return decryptedToken, nil
}

// set using the new info

func (ti *TokenKeyInfo) Set(key, value string) error {
	kmsService, err := cloudkms.NewService(context.TODO(), option.WithScopes(cloudkms.CloudPlatformScope))
	if err != nil {
		return errors.Errorf("error creating google kms service client: %s", err.Error())
	}

	name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		ti.opt.NewKmsProject, ti.opt.NewKmsLocation,
		ti.opt.NewKmsKeyRing, ti.opt.NewKmsCryptoKey)

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

	w := ti.storageClient.Bucket(ti.opt.NewBucket).Object(key).NewWriter(context.TODO())
	if _, err := w.Write(cipherText); err != nil {
		return fmt.Errorf("error writing key '%s' to gcs bucket '%s'", key, ti.opt.NewBucket)
	}

	return w.Close()
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

func (ti *TokenKeyInfo) TokenName() string {
	sts, err := ti.kubeClient.AppsV1().StatefulSets(ti.opt.AppBindingNamespace).Get(context.TODO(), ti.opt.AppBindingName, metav1.GetOptions{})
	if err != nil {
		return ""
	}

	var keyPrefix string
	for _, cont := range sts.Spec.Template.Spec.Containers {
		if cont.Name != vaultapi.VaultUnsealerContainerName {
			continue
		}
		for _, arg := range cont.Args {
			if strings.HasPrefix(arg, "--key-prefix=") {
				keyPrefix = arg[1+strings.Index(arg, "="):]
			}
		}
	}

	return fmt.Sprintf("%s-root-token", keyPrefix)
}

func (ti *TokenKeyInfo) UnsealKeyName(id int) (string, error) {
	sts, err := ti.kubeClient.AppsV1().StatefulSets(ti.opt.AppBindingNamespace).Get(context.TODO(), ti.opt.AppBindingNamespace, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	var keyPrefix string
	for _, cont := range sts.Spec.Template.Spec.Containers {
		if cont.Name != vaultapi.VaultUnsealerContainerName {
			continue
		}
		for _, arg := range cont.Args {
			if strings.HasPrefix(arg, "--key-prefix=") {
				keyPrefix = arg[1+strings.Index(arg, "="):]
			}
		}
	}

	return fmt.Sprintf("%s-unseal-key-%d", keyPrefix, id), nil
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
