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
	"github.com/pkg/errors"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

type StoreInterface interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

func (opt *VaultOptions) newStore(vs *vaultapi.VaultServer) (StoreInterface, error) {
	if vs == nil {
		return nil, errors.New("vaultserver is nil")
	}

	mode := vs.Spec.Unsealer.Mode
	switch true {

	case mode.GoogleKmsGcs != nil:
		return opt.newGcsStore(vs)
	case mode.AwsKmsSsm != nil:
		return opt.newAwsKmsStore(vs)
	case mode.AzureKeyVault != nil:
		return opt.newAzureStore(vs)
	case mode.KubernetesSecret != nil:
		return opt.newK8sStore(vs)
	}

	return nil, errors.New("unknown unseal mode")
}
