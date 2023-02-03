// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//

package objectpriority

import (
	"errors"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cli-utils/pkg/object"
)

const (
	Annotation         = "config.kubernetes.io/priority-level"
	MaxPriority uint64 = 1000000000
)

var (
	staticPriorities = map[schema.GroupKind]uint64{
		schema.GroupKind{Group: "", Kind: "Namespace"}:                                    MaxPriority + 1,
		schema.GroupKind{Group: "apiextensions.k8s.io", Kind: "CustomResourceDefinition"}: MaxPriority + 1,
	}
)

// GetStaticPriority returns the static priority for the object if exists.
// It returns the priority value and if static priority was found.
func GetStaticPriority(u *unstructured.Unstructured) (uint64, bool) {
	if u == nil {
		return 0, false
	}
	gvk := u.GroupVersionKind()
	pri, found := staticPriorities[gvk.GroupKind()]
	return pri, found
}

// HasAnnotation returns true if the config.kubernetes.io/priority-level annotation
// is present, false if not.
func HasAnnotation(u *unstructured.Unstructured) bool {
	if u == nil {
		return false
	}
	_, found := u.GetAnnotations()[Annotation]
	return found
}

// ReadAnnotation reads the priority-level annotation and parses the priority level into an unsigned int
func ReadAnnotation(u *unstructured.Unstructured) (uint64, error) {
	if u == nil {
		return 0, nil
	}
	priorityStr, found := u.GetAnnotations()[Annotation]
	if !found {
		return 0, nil
	}
	klog.V(5).Infof("priority-level annotation found for %s/%s: %q",
		u.GetNamespace(), u.GetName(), priorityStr)

	priority, err := strconv.ParseUint(priorityStr, 10, 64)
	if err != nil {
		return 0, object.InvalidAnnotationError{
			Annotation: Annotation,
			Cause:      err,
		}
	}

	if priority > MaxPriority {
		return 0, object.InvalidAnnotationError{
			Annotation: Annotation,
			Cause:      fmt.Errorf("priority higher than the maximum allowed: %d", MaxPriority),
		}
	}

	return priority, nil
}

// WriteAnnotation updates the supplied unstructured object to add the
// priority-level annotation
func WriteAnnotation(obj *unstructured.Unstructured, priority uint64) error {
	if obj == nil {
		return errors.New("object is nil")
	}
	a := obj.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}
	a[Annotation] = fmt.Sprintf("%d", priority)
	obj.SetAnnotations(a)
	return nil
}
