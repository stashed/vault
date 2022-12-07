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

package store

import (
	"stash.appscode.dev/vault/pkg/store/aws"
	"stash.appscode.dev/vault/pkg/store/azure"
	"stash.appscode.dev/vault/pkg/store/gcs"
	"stash.appscode.dev/vault/pkg/store/k8s"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

func NewStore(kc kubernetes.Interface, appBinding *appcatalog.AppBinding, unsealerSpec *vaultapi.UnsealerSpec) (StoreInterface, error) {
	if appBinding == nil {
		return nil, errors.New("appBinding is nil")
	}

	if unsealerSpec == nil {
		return nil, errors.New("vault unsealer spec is nil")
	}

	if kc == nil {
		return nil, errors.New("kubeclient is nil")
	}

	mode := unsealerSpec.Mode
	switch true {

	case mode.GoogleKmsGcs != nil:
		return gcs.New(kc, appBinding, mode.GoogleKmsGcs)
	case mode.AwsKmsSsm != nil:
		return aws.New(kc, appBinding, mode.AwsKmsSsm)
	case mode.AzureKeyVault != nil:
		return azure.New(kc, appBinding, mode.AzureKeyVault)
	case mode.KubernetesSecret != nil:
		return k8s.New(kc, appBinding, mode.KubernetesSecret)
	}

	return nil, errors.New("unknown unseal mode")
}
