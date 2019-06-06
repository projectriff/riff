package kcache

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func listResourceVersion(obj runtime.Object) (string, error) {
	list, err := meta.ListAccessor(obj)
	if err != nil {
		return "", err
	}
	return list.GetResourceVersion(), nil
}

func extractList(obj runtime.Object) ([]metav1.Object, error) {
	olist, err := meta.ExtractList(obj)
	if err != nil {
		return nil, err
	}

	mlist := make([]metav1.Object, 0, len(olist))

	for _, obj := range olist {
		if obj, ok := obj.(metav1.Object); ok {
			mlist = append(mlist, obj)
			continue
		}
		return nil, errInvalidType
	}

	return mlist, nil
}
