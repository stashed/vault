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

package aws

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	AWSAccessKey = "AWS_ACCESS_KEY_ID"
	AWSSecretKey = "AWS_SECRET_ACCESS_KEY"
)

type AwsKmsStore struct {
	ssmService *ssm.SSM
	kmsService *kms.KMS
	vs         *vaultapi.VaultServer
}

func New(kc kubernetes.Interface, vs *vaultapi.VaultServer) (*AwsKmsStore, error) {
	if vs == nil {
		return nil, errors.New("vault server is nil")
	}

	if kc == nil {
		return nil, errors.New("kubeClient is nil")
	}

	var cred string
	if vs.Spec.Unsealer.Mode.AwsKmsSsm.CredentialSecretRef != nil {
		cred = vs.Spec.Unsealer.Mode.AwsKmsSsm.CredentialSecretRef.Name
	}

	secret, err := kc.CoreV1().Secrets(vs.Namespace).Get(context.TODO(), cred, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	accessKey, ok := secret.Data["access_key"]
	if ok {
		if err = os.Setenv(AWSAccessKey, string(accessKey)); err != nil {
			return nil, err
		}
	}

	secretKey, ok := secret.Data["secret_key"]
	if ok {
		if err = os.Setenv(AWSSecretKey, string(secretKey)); err != nil {
			return nil, err
		}
	}

	sess, err := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: func() *bool {
			f := true
			return &f
		}(),
		Region: aws.String(vs.Spec.Unsealer.Mode.AwsKmsSsm.Region),
	},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	return &AwsKmsStore{
		kmsService: kms.New(sess),
		ssmService: ssm.New(sess),
		vs:         vs,
	}, nil
}

func (store *AwsKmsStore) Get(key string) (string, error) {
	req := &ssm.GetParametersInput{
		Names: []*string{
			aws.String(key),
		},
		WithDecryption: aws.Bool(false),
	}

	params, err := store.ssmService.GetParameters(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to get key from ssm")
	}

	if len(params.Parameters) == 0 {
		return "", errors.New("failed to get key from ssm; empty response")
	}

	sDec, err := base64.StdEncoding.DecodeString(*params.Parameters[0].Value)
	if err != nil {
		return "", errors.Wrap(err, "failed to base64-decode")
	}

	awsKmsSsmSpec := store.vs.Spec.Unsealer.Mode.AwsKmsSsm
	decryptOutput, err := store.kmsService.Decrypt(&kms.DecryptInput{
		CiphertextBlob: sDec,
		EncryptionContext: map[string]*string{
			"Tool": aws.String("vault-unsealer"),
		},
		GrantTokens: []*string{},
		KeyId:       aws.String(awsKmsSsmSpec.KmsKeyID),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to kms decrypt")
	}

	return string(decryptOutput.Plaintext), nil
}

func (store *AwsKmsStore) Set(key, value string) error {
	awsKmsSsmSpec := store.vs.Spec.Unsealer.Mode.AwsKmsSsm

	out, err := store.kmsService.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(awsKmsSsmSpec.KmsKeyID),
		Plaintext: []byte(value),
		EncryptionContext: map[string]*string{
			"Tool": aws.String("vault-unsealer"),
		},
		GrantTokens: []*string{},
	})
	if err != nil {
		return err
	}

	req := &ssm.PutParameterInput{
		Description: aws.String("vault-unsealer"),
		Name:        aws.String(key),
		Overwrite:   aws.Bool(true),
		Type:        aws.String("String"),
		Value:       aws.String(base64.StdEncoding.EncodeToString(out.CiphertextBlob)),
	}

	_, err = store.ssmService.PutParameter(req)
	return err
}
