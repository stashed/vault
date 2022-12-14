/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"

	v1alpha1 "go.bytebuilders.dev/license-proxyserver/apis/proxyserver/v1alpha1"
	scheme "go.bytebuilders.dev/license-proxyserver/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rest "k8s.io/client-go/rest"
)

// LicenseRequestsGetter has a method to return a LicenseRequestInterface.
// A group's client should implement this interface.
type LicenseRequestsGetter interface {
	LicenseRequests() LicenseRequestInterface
}

// LicenseRequestInterface has methods to work with LicenseRequest resources.
type LicenseRequestInterface interface {
	Create(ctx context.Context, licenseRequest *v1alpha1.LicenseRequest, opts v1.CreateOptions) (*v1alpha1.LicenseRequest, error)
	LicenseRequestExpansion
}

// licenseRequests implements LicenseRequestInterface
type licenseRequests struct {
	client rest.Interface
}

// newLicenseRequests returns a LicenseRequests
func newLicenseRequests(c *ProxyserverV1alpha1Client) *licenseRequests {
	return &licenseRequests{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a licenseRequest and creates it.  Returns the server's representation of the licenseRequest, and an error, if there is any.
func (c *licenseRequests) Create(ctx context.Context, licenseRequest *v1alpha1.LicenseRequest, opts v1.CreateOptions) (result *v1alpha1.LicenseRequest, err error) {
	result = &v1alpha1.LicenseRequest{}
	err = c.client.Post().
		Resource("licenserequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(licenseRequest).
		Do(ctx).
		Into(result)
	return
}
