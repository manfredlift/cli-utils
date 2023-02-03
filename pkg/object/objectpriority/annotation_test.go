// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//

package objectpriority

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var cm = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "unused",
			"namespace": "unused",
			"annotations": map[string]interface{}{
				Annotation: "1",
			},
		},
	},
}

var ns = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": "unused",
		},
	},
}

var crd = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "CustomResourceDefinition",
		"metadata": map[string]interface{}{
			"name": "unused",
		},
	},
}

var noAnnotations = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "unused",
			"namespace": "unused",
		},
	},
}

var negativeAnnotation = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "unused",
			"namespace": "unused",
			"annotations": map[string]interface{}{
				Annotation: "-1",
			},
		},
	},
}

var tooLargeAnnotation = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "unused",
			"namespace": "unused",
			"annotations": map[string]interface{}{
				Annotation: "1000000001",
			},
		},
	},
}

func TestReadAnnotation(t *testing.T) {
	testCases := map[string]struct {
		obj      *unstructured.Unstructured
		expected uint64
		isError  bool
	}{
		"nil object is not found": {
			obj:      nil,
			expected: 0,
		},
		"Object with no annotations returns 0": {
			obj:      noAnnotations,
			expected: 0,
		},
		"Negative priority annotation returns an error": {
			obj:      negativeAnnotation,
			expected: 0,
			isError:  true,
		},
		"Too large priority returns an error": {
			obj:      tooLargeAnnotation,
			expected: 0,
			isError:  true,
		},
		"Annotation with positive value returns the parsed value": {
			obj:      cm,
			expected: 1,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			actual, err := ReadAnnotation(tc.obj)
			if tc.isError {
				if err == nil {
					t.Fatalf("expected error not received")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error received: %s", err)
				}
				if actual != tc.expected {
					t.Errorf("expected (%d), got (%d)", tc.expected, actual)
				}
			}
		})
	}
}

func TestGetStaticPriority(t *testing.T) {
	testCases := map[string]struct {
		obj           *unstructured.Unstructured
		expected      uint64
		expectedFound bool
	}{
		"nil object is not found": {
			obj:           nil,
			expected:      0,
			expectedFound: false,
		},
		"configmap returns no static priority": {
			obj:           cm,
			expected:      0,
			expectedFound: false,
		},
		"Namespace returns static priority": {
			obj:           ns,
			expected:      1000000001,
			expectedFound: true,
		},
		"CRD returns static priority": {
			obj:           crd,
			expected:      1000000001,
			expectedFound: true,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			actual, found := GetStaticPriority(tc.obj)

			if actual != tc.expected {
				t.Errorf("expected (%d), got (%d)", tc.expected, actual)
			}

			if found != tc.expectedFound {
				t.Errorf("expectedFound (%d), got (%d)", tc.expected, actual)
			}
		})
	}
}
