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
	"math/rand"
	"time"

	"stash.appscode.dev/vault/pkg/token-keys-store/api"

	"cloud.google.com/go/storage"
	"k8s.io/client-go/kubernetes"
)

const (
	ServiceAccountJSON    = "sa.json"
	GoogleApplicationCred = "GOOGLE_APPLICATION_CREDENTIALS"
)

type TokenKeyInfo struct {
	storageClient *storage.Client
	kubeClient    kubernetes.Interface
	path          string
}

var _ api.TokenKeyInterface = &TokenKeyInfo{}

func New(kubeClient kubernetes.Interface) (*TokenKeyInfo, error) {
	return nil, nil
}

func (ti *TokenKeyInfo) Get() (string, error) {
	return "", nil
}

func decryptSymmetric(name string, ciphertext []byte) (string, error) {
	return "", nil
}

func (ti *TokenKeyInfo) TokenName() string {
	return ""
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
