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
	"fmt"

	"github.com/pkg/errors"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	UnsealModeGoogleKmsGcs     = "googleKmsGcs"
	UnsealModeKubernetesSecret = "kubernetesSecret"
	UnsealModeAwsKmsSsm        = "awsKmsSsm"
	UnsealModeAzureKeyVault    = "azureKeyVault"
)

func (opt *VaultOptions) getTokenKeys() (map[string]string, error) {
	switch opt.unsealMode {
	case UnsealModeGoogleKmsGcs:
		return opt.getGcsTokenKeys()
	case UnsealModeKubernetesSecret:
		return opt.getK8sTokenKeys()
	case UnsealModeAwsKmsSsm:
		return opt.getAwsTokenKeys()
	case UnsealModeAzureKeyVault:
		return opt.getAzureTokenKeys()
	}

	return nil, errors.New("unknown unseal mode")
}

func (opt *VaultOptions) setTokenKeys(vs *vaultapi.VaultServer, keys map[string]string) error {
	mode := vs.Spec.Unsealer.Mode
	switch true {
	case mode.GoogleKmsGcs != nil:
		return opt.setGcsTokenKeys(vs, keys)
	case mode.KubernetesSecret != nil:
		return opt.setK8sTokenKeys(vs, keys)
	case mode.AwsKmsSsm != nil:
		return opt.setAwsTokenKeys(vs, keys)
	case mode.AzureKeyVault != nil:
		return opt.setAzureTokenKeys(vs, keys)
	}

	return errors.New("unknown unseal mode")
}

func (opt *VaultOptions) getKeys() map[string]string {
	keys := make(map[string]string)

	var key string
	key = opt.tokenName()
	keys[key] = ""
	for id := 0; int64(id) < opt.secretShares; id++ {
		key = opt.unsealKeyName(id)
		keys[key] = ""
	}

	return keys
}

func (opt *VaultOptions) unsealKeyName(id int) string {
	return fmt.Sprintf("%s-unseal-key-%d", opt.KeyPrefix, id)
}

func (opt *VaultOptions) tokenName() string {
	return fmt.Sprintf("%s-root-token", opt.KeyPrefix)
}
