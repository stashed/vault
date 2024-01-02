/*
Copyright AppsCode Inc. and Contributors

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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// MetricsConfigurationSpecApplyConfiguration represents an declarative configuration of the MetricsConfigurationSpec type for use
// with apply.
type MetricsConfigurationSpecApplyConfiguration struct {
	TargetRef    *TargetRefApplyConfiguration `json:"targetRef,omitempty"`
	CommonLabels []LabelApplyConfiguration    `json:"commonLabels,omitempty"`
	Metrics      []MetricsApplyConfiguration  `json:"metrics,omitempty"`
}

// MetricsConfigurationSpecApplyConfiguration constructs an declarative configuration of the MetricsConfigurationSpec type for use with
// apply.
func MetricsConfigurationSpec() *MetricsConfigurationSpecApplyConfiguration {
	return &MetricsConfigurationSpecApplyConfiguration{}
}

// WithTargetRef sets the TargetRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TargetRef field is set to the value of the last call.
func (b *MetricsConfigurationSpecApplyConfiguration) WithTargetRef(value *TargetRefApplyConfiguration) *MetricsConfigurationSpecApplyConfiguration {
	b.TargetRef = value
	return b
}

// WithCommonLabels adds the given value to the CommonLabels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the CommonLabels field.
func (b *MetricsConfigurationSpecApplyConfiguration) WithCommonLabels(values ...*LabelApplyConfiguration) *MetricsConfigurationSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithCommonLabels")
		}
		b.CommonLabels = append(b.CommonLabels, *values[i])
	}
	return b
}

// WithMetrics adds the given value to the Metrics field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Metrics field.
func (b *MetricsConfigurationSpecApplyConfiguration) WithMetrics(values ...*MetricsApplyConfiguration) *MetricsConfigurationSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithMetrics")
		}
		b.Metrics = append(b.Metrics, *values[i])
	}
	return b
}
