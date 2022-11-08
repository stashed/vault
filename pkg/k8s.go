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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

func (opt *VaultOptions) getK8sTokenKeys() (map[string]string, error) {
	secret, err := opt.kubeClient.CoreV1().Secrets(opt.appBindingNamespace).Get(context.TODO(), opt.secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	keys := opt.getKeys()
	for k, v := range secret.Data {
		keys[k] = string(v)
	}

	return keys, nil
}

func (opt *VaultOptions) setK8sTokenKeys(vs *vaultapi.VaultServer, keys map[string]string) error {
	mode := vs.Spec.Unsealer.Mode
	var secretName string
	if mode.KubernetesSecret != nil {
		secretName = mode.KubernetesSecret.SecretName
	}

	secret, err := opt.kubeClient.CoreV1().Secrets(opt.appBindingNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for key, value := range keys {
		secret.Data[key] = []byte(value)
	}

	_, err = opt.kubeClient.CoreV1().Secrets(opt.appBindingNamespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	return err
}
