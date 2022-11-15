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

package azure

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	AzureClientID     = "AZURE_CLIENT_ID"
	AzureClientSecret = "AZURE_CLIENT_SECRET"
	AzureTenantID     = "AZURE_TENANT_ID"
)

type AzureStore struct {
	cred *azidentity.DefaultAzureCredential
	vs   *vaultapi.VaultServer
}

func New(kc kubernetes.Interface, vs *vaultapi.VaultServer) (*AzureStore, error) {
	if vs == nil {
		return nil, errors.New("vault server is nil")
	}

	if kc == nil {
		return nil, errors.New("kubeClient is nil")
	}

	var cred string
	if vs.Spec.Unsealer.Mode.AzureKeyVault.CredentialSecretRef != nil {
		cred = vs.Spec.Unsealer.Mode.AzureKeyVault.CredentialSecretRef.Name
	}
	secret, err := kc.CoreV1().Secrets(vs.Namespace).Get(context.TODO(), cred, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	clientID, ok := secret.Data["client-id"]
	if ok {
		if err = os.Setenv(AzureClientID, string(clientID)); err != nil {
			return nil, err
		}
	}

	clientSecret, ok := secret.Data["client-secret"]
	if ok {
		if err = os.Setenv(AzureClientSecret, string(clientSecret)); err != nil {
			return nil, err
		}
	}

	if err := os.Setenv(AzureTenantID, vs.Spec.Unsealer.Mode.AzureKeyVault.TenantID); err != nil {
		return nil, err
	}

	azcred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	return &AzureStore{
		cred: azcred,
		vs:   vs,
	}, nil
}

func (store *AzureStore) Get(key string) (string, error) {
	vaultBaseUrl := store.vs.Spec.Unsealer.Mode.AzureKeyVault.VaultBaseURL
	client := azsecrets.NewClient(vaultBaseUrl, store.cred, nil)

	resp, err := client.GetSecret(context.Background(), strings.Replace(key, ".", "-", -1), "", nil)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(*resp.Value)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func (store *AzureStore) Set(key, value string) error {
	key = strings.Replace(key, ".", "-", -1)

	vaultBaseUrl := store.vs.Spec.Unsealer.Mode.AzureKeyVault.VaultBaseURL
	client := azsecrets.NewClient(vaultBaseUrl, store.cred, nil)

	_, err := client.SetSecret(context.TODO(), key, azsecrets.SetSecretParameters{
		Value:       pointer.StringP(base64.StdEncoding.EncodeToString([]byte(value))),
		ContentType: pointer.StringP("password"),
	}, nil)
	if err != nil {
		return errors.Wrap(err, "unable to set secrets in key vault")
	}

	return nil
}
