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

package k8s

import (
	"context"

	"github.com/pkg/errors"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

type K8sStore struct {
	k8sSpec    *vaultapi.KubernetesSecretSpec
	kc         kubernetes.Interface
	appBinding *appcatalog.AppBinding
}

func New(kc kubernetes.Interface, appBinding *appcatalog.AppBinding, k8sSpec *vaultapi.KubernetesSecretSpec) (*K8sStore, error) {
	if k8sSpec == nil {
		return nil, errors.New("k8sSpec  is nil")
	}

	if kc == nil {
		return nil, errors.New("kubeClient is nil")
	}

	return &K8sStore{
		k8sSpec:    k8sSpec,
		kc:         kc,
		appBinding: appBinding,
	}, nil
}

func (store *K8sStore) Get(key string) (string, error) {
	name := store.k8sSpec.SecretName

	secret, err := store.kc.CoreV1().Secrets(store.appBinding.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if _, ok := secret.Data[key]; !ok {
		return "", errors.Errorf("%s not found in secret %s/%s", key, store.appBinding.Namespace, name)
	}

	return string(secret.Data[key]), nil
}

func (store *K8sStore) Set(key, value string) error {
	name := store.k8sSpec.SecretName
	secret, err := store.kc.CoreV1().Secrets(store.appBinding.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors2.IsNotFound(err) {
			return nil
		}
		return err
	}

	secret.Data[key] = []byte(value)

	_, err = store.kc.CoreV1().Secrets(store.appBinding.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	return err
}
