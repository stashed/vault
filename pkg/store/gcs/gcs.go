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

package gcs

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

	kmsv1 "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	ServiceAccountJSON    = "sa.json"
	GoogleApplicationCred = "GOOGLE_APPLICATION_CREDENTIALS"
)

type gcsStore struct {
	gcsSpec    *vaultapi.GoogleKmsGcsSpec
	client     *storage.Client
	appBinding *appcatalog.AppBinding
}

func New(kc kubernetes.Interface, appBinding *appcatalog.AppBinding, gcsSpec *vaultapi.GoogleKmsGcsSpec) (*gcsStore, error) {
	if appBinding == nil {
		return nil, fmt.Errorf("appBinding is nil")
	}

	if gcsSpec == nil {
		return nil, fmt.Errorf("gcsSpec is nil")
	}

	if kc == nil {
		return nil, fmt.Errorf("kubeClient is nil")
	}

	var cred string
	if gcsSpec.CredentialSecretRef != nil {
		cred = gcsSpec.CredentialSecretRef.Name
	}

	secret, err := kc.CoreV1().Secrets(appBinding.Namespace).Get(context.TODO(), cred, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if _, ok := secret.Data[ServiceAccountJSON]; !ok {
		return nil, fmt.Errorf("%s not found in secret", ServiceAccountJSON)
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

	return &gcsStore{
		gcsSpec:    gcsSpec,
		client:     client,
		appBinding: appBinding,
	}, nil
}

func (store *gcsStore) Get(key string) (string, error) {
	rc, err := store.client.Bucket(store.gcsSpec.Bucket).Object(key).NewReader(context.TODO())
	if err != nil {
		return "", err
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		store.gcsSpec.KmsProject, store.gcsSpec.KmsLocation,
		store.gcsSpec.KmsKeyRing, store.gcsSpec.KmsCryptoKey)

	decryptedToken, err := decryptSymmetric(name, body)
	if err != nil {
		return "", err
	}

	return decryptedToken, nil
}

func (store *gcsStore) Set(key, value string) error {
	kmsService, err := cloudkms.NewService(context.TODO(), option.WithScopes(cloudkms.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("error creating google kms service client: %w", err)
	}

	name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		store.gcsSpec.KmsProject, store.gcsSpec.KmsLocation,
		store.gcsSpec.KmsKeyRing, store.gcsSpec.KmsCryptoKey)

	resp, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.Encrypt(name, &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString([]byte(value)),
	}).Do()
	if err != nil {
		return fmt.Errorf("error encrypting data: %w", err)
	}

	cipherText, err := base64.StdEncoding.DecodeString(resp.Ciphertext)
	if err != nil {
		return err
	}

	w := store.client.Bucket(store.gcsSpec.Bucket).Object(key).NewWriter(context.TODO())
	if _, err := w.Write(cipherText); err != nil {
		return fmt.Errorf("error writing key %s to gcs bucket %s: %w", key, store.gcsSpec.Bucket, err)
	}

	return w.Close()
}

func decryptSymmetric(name string, ciphertext []byte) (string, error) {
	client, err := kmsv1.NewKeyManagementClient(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to create kms client: %w", err)
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
		return "", fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	if int64(crc32c(result.Plaintext)) != result.PlaintextCrc32C.Value {
		return "", fmt.Errorf("decrypt response corrupted in-transit")
	}

	return string(result.Plaintext), nil
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
