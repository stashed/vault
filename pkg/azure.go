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
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	AzureClientID     = "AZURE_CLIENT_ID"
	AzureClientSecret = "AZURE_CLIENT_SECRET"
	AzureTenantID     = "AZURE_TENANT_ID"
)

func (opt *VaultOptions) newAzureCred(cred, tenantID string) (*azidentity.DefaultAzureCredential, error) {
	secret, err := opt.kubeClient.CoreV1().Secrets(opt.appBindingNamespace).Get(context.TODO(), cred, metav1.GetOptions{})
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

	if err = os.Setenv(AzureTenantID, tenantID); err != nil {
		return nil, err
	}

	azCred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	return azCred, nil
}

func (opt *VaultOptions) getAzureTokenKeys() (map[string]string, error) {
	azCred, err := opt.newAzureCred(opt.credentialSecretRef, opt.tenantID)
	if err != nil {
		return nil, err
	}

	client := azsecrets.NewClient(opt.vaultBaseURL, azCred, nil)
	keys := opt.getKeys()
	for key := range keys {
		resp, err := client.GetSecret(context.Background(), strings.Replace(key, ".", "-", -1), "", nil)
		if err != nil {
			return nil, err
		}

		decoded, err := base64.StdEncoding.DecodeString(*resp.Value)
		if err != nil {
			return nil, err
		}

		keys[key] = string(decoded)
	}

	return keys, nil
}

func (opt *VaultOptions) setAzureTokenKeys(vs *vaultapi.VaultServer, keys map[string]string) error {
	mode := vs.Spec.Unsealer.Mode

	var credRef string
	if mode.AzureKeyVault.CredentialSecretRef != nil {
		credRef = mode.AzureKeyVault.CredentialSecretRef.Name
	}

	azCred, err := opt.newAzureCred(credRef, mode.AzureKeyVault.TenantID)
	if err != nil {
		return err
	}

	for key, value := range keys {
		key = strings.Replace(key, ".", "-", -1)

		vaultBaseUrl := vs.Spec.Unsealer.Mode.AzureKeyVault.VaultBaseURL
		client := azsecrets.NewClient(vaultBaseUrl, azCred, nil)

		_, err := client.SetSecret(context.TODO(), key, azsecrets.SetSecretParameters{
			Value:       pointer.StringP(base64.StdEncoding.EncodeToString([]byte(value))),
			ContentType: pointer.StringP("password"),
		}, nil)
		if err != nil {
			return errors.Wrap(err, "unable to set secrets in key vault")
		}
	}

	return nil
}
