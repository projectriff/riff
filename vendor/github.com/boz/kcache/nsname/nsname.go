package nsname

import (
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrInvalidID = errors.New("Invalid ID")

type NSName struct {
	Namespace string
	Name      string
}

func Parse(id string) (NSName, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return NSName{}, ErrInvalidID
	}
	return New(parts[0], parts[1]), nil
}

func New(ns, name string) NSName {
	return NSName{ns, name}
}

func ForObject(obj metav1.Object) NSName {
	return New(obj.GetNamespace(), obj.GetName())
}

func (obj NSName) String() string {
	return fmt.Sprintf("%v/%v", obj.Namespace, obj.Name)
}
