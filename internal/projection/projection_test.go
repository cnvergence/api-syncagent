/*
Copyright 2024 The Kubermatic Kubernetes Platform contributors.

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

package projection

import (
	"testing"

	kdpservicesv1alpha1 "k8c.io/servlet/sdk/apis/services/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestPublishedResourceSourceGVK(t *testing.T) {
	const (
		apiGroup = "testgroup"
		version  = "v1"
		kind     = "test"
	)

	pubRes := kdpservicesv1alpha1.PublishedResource{
		Spec: kdpservicesv1alpha1.PublishedResourceSpec{
			Resource: kdpservicesv1alpha1.SourceResourceDescriptor{
				APIGroup: apiGroup,
				Version:  version,
				Kind:     kind,
			},
		},
	}

	gvk := PublishedResourceSourceGVK(&pubRes)

	if gvk.Group != apiGroup {
		t.Errorf("Expected API group to be %q, but got %q.", apiGroup, gvk.Group)
	}

	if gvk.Version != version {
		t.Errorf("Expected version to be %q, but got %q.", version, gvk.Version)
	}

	if gvk.Kind != kind {
		t.Errorf("Expected kind to be %q, but got %q.", kind, gvk.Kind)
	}
}

func TestPublishedResourceProjectedGVK(t *testing.T) {
	const (
		apiGroup         = "testgroup"
		overrideAPIGroup = "newgroup"
		version          = "v1"
		kind             = "test"
	)

	pubRes := &kdpservicesv1alpha1.PublishedResource{
		Spec: kdpservicesv1alpha1.PublishedResourceSpec{
			Resource: kdpservicesv1alpha1.SourceResourceDescriptor{
				APIGroup: apiGroup,
				Version:  version,
				Kind:     kind,
			},
		},
	}

	testcases := []struct {
		name       string
		projection *kdpservicesv1alpha1.ResourceProjection
		expected   schema.GroupVersionKind
	}{
		{
			name:       "no projection",
			projection: nil,
			expected:   schema.GroupVersionKind{Group: overrideAPIGroup, Version: version, Kind: kind},
		},
		{
			name:       "override version",
			projection: &kdpservicesv1alpha1.ResourceProjection{Version: "v2"},
			expected:   schema.GroupVersionKind{Group: overrideAPIGroup, Version: "v2", Kind: kind},
		},
		{
			name:       "override kind",
			projection: &kdpservicesv1alpha1.ResourceProjection{Kind: "dummy"},
			expected:   schema.GroupVersionKind{Group: overrideAPIGroup, Version: version, Kind: "dummy"},
		},
		{
			name:       "override both",
			projection: &kdpservicesv1alpha1.ResourceProjection{Version: "v2", Kind: "dummy"},
			expected:   schema.GroupVersionKind{Group: overrideAPIGroup, Version: "v2", Kind: "dummy"},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			pr := pubRes.DeepCopy()
			pr.Spec.Projection = testcase.projection

			gvk := PublishedResourceProjectedGVK(pr, overrideAPIGroup)

			if gvk.Group != testcase.expected.Group {
				t.Errorf("Expected API group to be %q, but got %q.", testcase.expected.Group, gvk.Group)
			}

			if gvk.Version != testcase.expected.Version {
				t.Errorf("Expected version to be %q, but got %q.", testcase.expected.Version, gvk.Version)
			}

			if gvk.Kind != testcase.expected.Kind {
				t.Errorf("Expected kind to be %q, but got %q.", testcase.expected.Kind, gvk.Kind)
			}
		})
	}
}
