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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha2"
)

const (
	AWSAccessKey = "AWS_ACCESS_KEY_ID"
	AWSSecretKey = "AWS_SECRET_ACCESS_KEY"
)

func (opt *VaultOptions) newAwsSession(cred, region string) (*session.Session, error) {
	secret, err := opt.kubeClient.CoreV1().Secrets(opt.appBindingNamespace).Get(context.TODO(), cred, metav1.GetOptions{})
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
		Region: aws.String(region),
	},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	return sess, nil
}

func (opt *VaultOptions) getAwsTokenKeys() (map[string]string, error) {
	sess, err := opt.newAwsSession(opt.credentialSecretRef, opt.region)
	if err != nil {
		return nil, err
	}

	kmsService := kms.New(sess)
	ssmService := ssm.New(sess)

	keys := opt.getKeys()
	for key := range keys {
		req := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(key),
			},
			WithDecryption: aws.Bool(false),
		}

		params, err := ssmService.GetParameters(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get key from ssm")
		}

		if len(params.Parameters) == 0 {
			return nil, errors.New("failed to get key from ssm; empty response")
		}

		// Since len of the params is greater than zero
		sDec, err := base64.StdEncoding.DecodeString(*params.Parameters[0].Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to base64-decode")
		}

		decryptOutput, err := kmsService.Decrypt(&kms.DecryptInput{
			CiphertextBlob: sDec,
			EncryptionContext: map[string]*string{
				"Tool": aws.String("vault-unsealer"),
			},
			GrantTokens: []*string{},
			KeyId:       aws.String(opt.kmsKeyID),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to kms decrypt")
		}

		keys[key] = string(decryptOutput.Plaintext)
	}

	return keys, nil
}

func (opt *VaultOptions) setAwsTokenKeys(vs *vaultapi.VaultServer, keys map[string]string) error {
	mode := vs.Spec.Unsealer.Mode

	var credRef string
	if mode.AwsKmsSsm.CredentialSecretRef != nil {
		credRef = mode.AwsKmsSsm.CredentialSecretRef.Name
	}

	sess, err := opt.newAwsSession(credRef, mode.AwsKmsSsm.Region)
	if err != nil {
		return err
	}

	kmsService := kms.New(sess)
	ssmService := ssm.New(sess)

	awsKmsSsmSpec := vs.Spec.Unsealer.Mode.AwsKmsSsm

	for key, value := range keys {
		out, err := kmsService.Encrypt(&kms.EncryptInput{
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

		if _, err = ssmService.PutParameter(req); err != nil {
			return err
		}
	}

	return nil
}
