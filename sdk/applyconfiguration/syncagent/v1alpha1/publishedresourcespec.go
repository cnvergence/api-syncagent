/*
Copyright 2025 The KCP Authors.

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

// PublishedResourceSpecApplyConfiguration represents a declarative configuration of the PublishedResourceSpec type for use
// with apply.
type PublishedResourceSpecApplyConfiguration struct {
	Resource             *SourceResourceDescriptorApplyConfiguration `json:"resource,omitempty"`
	Filter               *ResourceFilterApplyConfiguration           `json:"filter,omitempty"`
	Naming               *ResourceNamingApplyConfiguration           `json:"naming,omitempty"`
	EnableWorkspacePaths *bool                                       `json:"enableWorkspacePaths,omitempty"`
	Projection           *ResourceProjectionApplyConfiguration       `json:"projection,omitempty"`
	Mutation             *ResourceMutationSpecApplyConfiguration     `json:"mutation,omitempty"`
	Related              []RelatedResourceSpecApplyConfiguration     `json:"related,omitempty"`
}

// PublishedResourceSpecApplyConfiguration constructs a declarative configuration of the PublishedResourceSpec type for use with
// apply.
func PublishedResourceSpec() *PublishedResourceSpecApplyConfiguration {
	return &PublishedResourceSpecApplyConfiguration{}
}

// WithResource sets the Resource field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Resource field is set to the value of the last call.
func (b *PublishedResourceSpecApplyConfiguration) WithResource(value *SourceResourceDescriptorApplyConfiguration) *PublishedResourceSpecApplyConfiguration {
	b.Resource = value
	return b
}

// WithFilter sets the Filter field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Filter field is set to the value of the last call.
func (b *PublishedResourceSpecApplyConfiguration) WithFilter(value *ResourceFilterApplyConfiguration) *PublishedResourceSpecApplyConfiguration {
	b.Filter = value
	return b
}

// WithNaming sets the Naming field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Naming field is set to the value of the last call.
func (b *PublishedResourceSpecApplyConfiguration) WithNaming(value *ResourceNamingApplyConfiguration) *PublishedResourceSpecApplyConfiguration {
	b.Naming = value
	return b
}

// WithEnableWorkspacePaths sets the EnableWorkspacePaths field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the EnableWorkspacePaths field is set to the value of the last call.
func (b *PublishedResourceSpecApplyConfiguration) WithEnableWorkspacePaths(value bool) *PublishedResourceSpecApplyConfiguration {
	b.EnableWorkspacePaths = &value
	return b
}

// WithProjection sets the Projection field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Projection field is set to the value of the last call.
func (b *PublishedResourceSpecApplyConfiguration) WithProjection(value *ResourceProjectionApplyConfiguration) *PublishedResourceSpecApplyConfiguration {
	b.Projection = value
	return b
}

// WithMutation sets the Mutation field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Mutation field is set to the value of the last call.
func (b *PublishedResourceSpecApplyConfiguration) WithMutation(value *ResourceMutationSpecApplyConfiguration) *PublishedResourceSpecApplyConfiguration {
	b.Mutation = value
	return b
}

// WithRelated adds the given value to the Related field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Related field.
func (b *PublishedResourceSpecApplyConfiguration) WithRelated(values ...*RelatedResourceSpecApplyConfiguration) *PublishedResourceSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithRelated")
		}
		b.Related = append(b.Related, *values[i])
	}
	return b
}
