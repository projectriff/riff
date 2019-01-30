// Code generated by mockery v1.0.0. DO NOT EDIT.

package vendor_mocks

import appsv1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
import auditregistrationv1alpha1 "k8s.io/client-go/kubernetes/typed/auditregistration/v1alpha1"
import authenticationv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
import authenticationv1beta1 "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
import authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
import authorizationv1beta1 "k8s.io/client-go/kubernetes/typed/authorization/v1beta1"
import autoscalingv1 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
import batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
import batchv1beta1 "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
import certificatesv1beta1 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
import coordinationv1beta1 "k8s.io/client-go/kubernetes/typed/coordination/v1beta1"
import corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
import discovery "k8s.io/client-go/discovery"
import eventsv1beta1 "k8s.io/client-go/kubernetes/typed/events/v1beta1"
import extensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"

import mock "github.com/stretchr/testify/mock"
import networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"
import policyv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
import rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
import rbacv1alpha1 "k8s.io/client-go/kubernetes/typed/rbac/v1alpha1"
import rbacv1beta1 "k8s.io/client-go/kubernetes/typed/rbac/v1beta1"
import schedulingv1alpha1 "k8s.io/client-go/kubernetes/typed/scheduling/v1alpha1"
import schedulingv1beta1 "k8s.io/client-go/kubernetes/typed/scheduling/v1beta1"
import settingsv1alpha1 "k8s.io/client-go/kubernetes/typed/settings/v1alpha1"
import storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
import storagev1alpha1 "k8s.io/client-go/kubernetes/typed/storage/v1alpha1"
import storagev1beta1 "k8s.io/client-go/kubernetes/typed/storage/v1beta1"
import v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
import v1alpha1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1alpha1"
import v1beta1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
import v1beta2 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
import v2alpha1 "k8s.io/client-go/kubernetes/typed/batch/v2alpha1"
import v2beta1 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta1"
import v2beta2 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta2"

// Interface is an autogenerated mock type for the Interface type
type Interface struct {
	mock.Mock
}

// Admissionregistration provides a mock function with given fields:
func (_m *Interface) Admissionregistration() v1beta1.AdmissionregistrationV1beta1Interface {
	ret := _m.Called()

	var r0 v1beta1.AdmissionregistrationV1beta1Interface
	if rf, ok := ret.Get(0).(func() v1beta1.AdmissionregistrationV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1beta1.AdmissionregistrationV1beta1Interface)
		}
	}

	return r0
}

// AdmissionregistrationV1alpha1 provides a mock function with given fields:
func (_m *Interface) AdmissionregistrationV1alpha1() v1alpha1.AdmissionregistrationV1alpha1Interface {
	ret := _m.Called()

	var r0 v1alpha1.AdmissionregistrationV1alpha1Interface
	if rf, ok := ret.Get(0).(func() v1alpha1.AdmissionregistrationV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.AdmissionregistrationV1alpha1Interface)
		}
	}

	return r0
}

// AdmissionregistrationV1beta1 provides a mock function with given fields:
func (_m *Interface) AdmissionregistrationV1beta1() v1beta1.AdmissionregistrationV1beta1Interface {
	ret := _m.Called()

	var r0 v1beta1.AdmissionregistrationV1beta1Interface
	if rf, ok := ret.Get(0).(func() v1beta1.AdmissionregistrationV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1beta1.AdmissionregistrationV1beta1Interface)
		}
	}

	return r0
}

// Apps provides a mock function with given fields:
func (_m *Interface) Apps() v1.AppsV1Interface {
	ret := _m.Called()

	var r0 v1.AppsV1Interface
	if rf, ok := ret.Get(0).(func() v1.AppsV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1.AppsV1Interface)
		}
	}

	return r0
}

// AppsV1 provides a mock function with given fields:
func (_m *Interface) AppsV1() v1.AppsV1Interface {
	ret := _m.Called()

	var r0 v1.AppsV1Interface
	if rf, ok := ret.Get(0).(func() v1.AppsV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1.AppsV1Interface)
		}
	}

	return r0
}

// AppsV1beta1 provides a mock function with given fields:
func (_m *Interface) AppsV1beta1() appsv1beta1.AppsV1beta1Interface {
	ret := _m.Called()

	var r0 appsv1beta1.AppsV1beta1Interface
	if rf, ok := ret.Get(0).(func() appsv1beta1.AppsV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(appsv1beta1.AppsV1beta1Interface)
		}
	}

	return r0
}

// AppsV1beta2 provides a mock function with given fields:
func (_m *Interface) AppsV1beta2() v1beta2.AppsV1beta2Interface {
	ret := _m.Called()

	var r0 v1beta2.AppsV1beta2Interface
	if rf, ok := ret.Get(0).(func() v1beta2.AppsV1beta2Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1beta2.AppsV1beta2Interface)
		}
	}

	return r0
}

// Auditregistration provides a mock function with given fields:
func (_m *Interface) Auditregistration() auditregistrationv1alpha1.AuditregistrationV1alpha1Interface {
	ret := _m.Called()

	var r0 auditregistrationv1alpha1.AuditregistrationV1alpha1Interface
	if rf, ok := ret.Get(0).(func() auditregistrationv1alpha1.AuditregistrationV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(auditregistrationv1alpha1.AuditregistrationV1alpha1Interface)
		}
	}

	return r0
}

// AuditregistrationV1alpha1 provides a mock function with given fields:
func (_m *Interface) AuditregistrationV1alpha1() auditregistrationv1alpha1.AuditregistrationV1alpha1Interface {
	ret := _m.Called()

	var r0 auditregistrationv1alpha1.AuditregistrationV1alpha1Interface
	if rf, ok := ret.Get(0).(func() auditregistrationv1alpha1.AuditregistrationV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(auditregistrationv1alpha1.AuditregistrationV1alpha1Interface)
		}
	}

	return r0
}

// Authentication provides a mock function with given fields:
func (_m *Interface) Authentication() authenticationv1.AuthenticationV1Interface {
	ret := _m.Called()

	var r0 authenticationv1.AuthenticationV1Interface
	if rf, ok := ret.Get(0).(func() authenticationv1.AuthenticationV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authenticationv1.AuthenticationV1Interface)
		}
	}

	return r0
}

// AuthenticationV1 provides a mock function with given fields:
func (_m *Interface) AuthenticationV1() authenticationv1.AuthenticationV1Interface {
	ret := _m.Called()

	var r0 authenticationv1.AuthenticationV1Interface
	if rf, ok := ret.Get(0).(func() authenticationv1.AuthenticationV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authenticationv1.AuthenticationV1Interface)
		}
	}

	return r0
}

// AuthenticationV1beta1 provides a mock function with given fields:
func (_m *Interface) AuthenticationV1beta1() authenticationv1beta1.AuthenticationV1beta1Interface {
	ret := _m.Called()

	var r0 authenticationv1beta1.AuthenticationV1beta1Interface
	if rf, ok := ret.Get(0).(func() authenticationv1beta1.AuthenticationV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authenticationv1beta1.AuthenticationV1beta1Interface)
		}
	}

	return r0
}

// Authorization provides a mock function with given fields:
func (_m *Interface) Authorization() authorizationv1.AuthorizationV1Interface {
	ret := _m.Called()

	var r0 authorizationv1.AuthorizationV1Interface
	if rf, ok := ret.Get(0).(func() authorizationv1.AuthorizationV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authorizationv1.AuthorizationV1Interface)
		}
	}

	return r0
}

// AuthorizationV1 provides a mock function with given fields:
func (_m *Interface) AuthorizationV1() authorizationv1.AuthorizationV1Interface {
	ret := _m.Called()

	var r0 authorizationv1.AuthorizationV1Interface
	if rf, ok := ret.Get(0).(func() authorizationv1.AuthorizationV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authorizationv1.AuthorizationV1Interface)
		}
	}

	return r0
}

// AuthorizationV1beta1 provides a mock function with given fields:
func (_m *Interface) AuthorizationV1beta1() authorizationv1beta1.AuthorizationV1beta1Interface {
	ret := _m.Called()

	var r0 authorizationv1beta1.AuthorizationV1beta1Interface
	if rf, ok := ret.Get(0).(func() authorizationv1beta1.AuthorizationV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authorizationv1beta1.AuthorizationV1beta1Interface)
		}
	}

	return r0
}

// Autoscaling provides a mock function with given fields:
func (_m *Interface) Autoscaling() autoscalingv1.AutoscalingV1Interface {
	ret := _m.Called()

	var r0 autoscalingv1.AutoscalingV1Interface
	if rf, ok := ret.Get(0).(func() autoscalingv1.AutoscalingV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(autoscalingv1.AutoscalingV1Interface)
		}
	}

	return r0
}

// AutoscalingV1 provides a mock function with given fields:
func (_m *Interface) AutoscalingV1() autoscalingv1.AutoscalingV1Interface {
	ret := _m.Called()

	var r0 autoscalingv1.AutoscalingV1Interface
	if rf, ok := ret.Get(0).(func() autoscalingv1.AutoscalingV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(autoscalingv1.AutoscalingV1Interface)
		}
	}

	return r0
}

// AutoscalingV2beta1 provides a mock function with given fields:
func (_m *Interface) AutoscalingV2beta1() v2beta1.AutoscalingV2beta1Interface {
	ret := _m.Called()

	var r0 v2beta1.AutoscalingV2beta1Interface
	if rf, ok := ret.Get(0).(func() v2beta1.AutoscalingV2beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v2beta1.AutoscalingV2beta1Interface)
		}
	}

	return r0
}

// AutoscalingV2beta2 provides a mock function with given fields:
func (_m *Interface) AutoscalingV2beta2() v2beta2.AutoscalingV2beta2Interface {
	ret := _m.Called()

	var r0 v2beta2.AutoscalingV2beta2Interface
	if rf, ok := ret.Get(0).(func() v2beta2.AutoscalingV2beta2Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v2beta2.AutoscalingV2beta2Interface)
		}
	}

	return r0
}

// Batch provides a mock function with given fields:
func (_m *Interface) Batch() batchv1.BatchV1Interface {
	ret := _m.Called()

	var r0 batchv1.BatchV1Interface
	if rf, ok := ret.Get(0).(func() batchv1.BatchV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(batchv1.BatchV1Interface)
		}
	}

	return r0
}

// BatchV1 provides a mock function with given fields:
func (_m *Interface) BatchV1() batchv1.BatchV1Interface {
	ret := _m.Called()

	var r0 batchv1.BatchV1Interface
	if rf, ok := ret.Get(0).(func() batchv1.BatchV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(batchv1.BatchV1Interface)
		}
	}

	return r0
}

// BatchV1beta1 provides a mock function with given fields:
func (_m *Interface) BatchV1beta1() batchv1beta1.BatchV1beta1Interface {
	ret := _m.Called()

	var r0 batchv1beta1.BatchV1beta1Interface
	if rf, ok := ret.Get(0).(func() batchv1beta1.BatchV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(batchv1beta1.BatchV1beta1Interface)
		}
	}

	return r0
}

// BatchV2alpha1 provides a mock function with given fields:
func (_m *Interface) BatchV2alpha1() v2alpha1.BatchV2alpha1Interface {
	ret := _m.Called()

	var r0 v2alpha1.BatchV2alpha1Interface
	if rf, ok := ret.Get(0).(func() v2alpha1.BatchV2alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v2alpha1.BatchV2alpha1Interface)
		}
	}

	return r0
}

// Certificates provides a mock function with given fields:
func (_m *Interface) Certificates() certificatesv1beta1.CertificatesV1beta1Interface {
	ret := _m.Called()

	var r0 certificatesv1beta1.CertificatesV1beta1Interface
	if rf, ok := ret.Get(0).(func() certificatesv1beta1.CertificatesV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(certificatesv1beta1.CertificatesV1beta1Interface)
		}
	}

	return r0
}

// CertificatesV1beta1 provides a mock function with given fields:
func (_m *Interface) CertificatesV1beta1() certificatesv1beta1.CertificatesV1beta1Interface {
	ret := _m.Called()

	var r0 certificatesv1beta1.CertificatesV1beta1Interface
	if rf, ok := ret.Get(0).(func() certificatesv1beta1.CertificatesV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(certificatesv1beta1.CertificatesV1beta1Interface)
		}
	}

	return r0
}

// Coordination provides a mock function with given fields:
func (_m *Interface) Coordination() coordinationv1beta1.CoordinationV1beta1Interface {
	ret := _m.Called()

	var r0 coordinationv1beta1.CoordinationV1beta1Interface
	if rf, ok := ret.Get(0).(func() coordinationv1beta1.CoordinationV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(coordinationv1beta1.CoordinationV1beta1Interface)
		}
	}

	return r0
}

// CoordinationV1beta1 provides a mock function with given fields:
func (_m *Interface) CoordinationV1beta1() coordinationv1beta1.CoordinationV1beta1Interface {
	ret := _m.Called()

	var r0 coordinationv1beta1.CoordinationV1beta1Interface
	if rf, ok := ret.Get(0).(func() coordinationv1beta1.CoordinationV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(coordinationv1beta1.CoordinationV1beta1Interface)
		}
	}

	return r0
}

// Core provides a mock function with given fields:
func (_m *Interface) Core() corev1.CoreV1Interface {
	ret := _m.Called()

	var r0 corev1.CoreV1Interface
	if rf, ok := ret.Get(0).(func() corev1.CoreV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(corev1.CoreV1Interface)
		}
	}

	return r0
}

// CoreV1 provides a mock function with given fields:
func (_m *Interface) CoreV1() corev1.CoreV1Interface {
	ret := _m.Called()

	var r0 corev1.CoreV1Interface
	if rf, ok := ret.Get(0).(func() corev1.CoreV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(corev1.CoreV1Interface)
		}
	}

	return r0
}

// Discovery provides a mock function with given fields:
func (_m *Interface) Discovery() discovery.DiscoveryInterface {
	ret := _m.Called()

	var r0 discovery.DiscoveryInterface
	if rf, ok := ret.Get(0).(func() discovery.DiscoveryInterface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(discovery.DiscoveryInterface)
		}
	}

	return r0
}

// Events provides a mock function with given fields:
func (_m *Interface) Events() eventsv1beta1.EventsV1beta1Interface {
	ret := _m.Called()

	var r0 eventsv1beta1.EventsV1beta1Interface
	if rf, ok := ret.Get(0).(func() eventsv1beta1.EventsV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(eventsv1beta1.EventsV1beta1Interface)
		}
	}

	return r0
}

// EventsV1beta1 provides a mock function with given fields:
func (_m *Interface) EventsV1beta1() eventsv1beta1.EventsV1beta1Interface {
	ret := _m.Called()

	var r0 eventsv1beta1.EventsV1beta1Interface
	if rf, ok := ret.Get(0).(func() eventsv1beta1.EventsV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(eventsv1beta1.EventsV1beta1Interface)
		}
	}

	return r0
}

// Extensions provides a mock function with given fields:
func (_m *Interface) Extensions() extensionsv1beta1.ExtensionsV1beta1Interface {
	ret := _m.Called()

	var r0 extensionsv1beta1.ExtensionsV1beta1Interface
	if rf, ok := ret.Get(0).(func() extensionsv1beta1.ExtensionsV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(extensionsv1beta1.ExtensionsV1beta1Interface)
		}
	}

	return r0
}

// ExtensionsV1beta1 provides a mock function with given fields:
func (_m *Interface) ExtensionsV1beta1() extensionsv1beta1.ExtensionsV1beta1Interface {
	ret := _m.Called()

	var r0 extensionsv1beta1.ExtensionsV1beta1Interface
	if rf, ok := ret.Get(0).(func() extensionsv1beta1.ExtensionsV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(extensionsv1beta1.ExtensionsV1beta1Interface)
		}
	}

	return r0
}

// Networking provides a mock function with given fields:
func (_m *Interface) Networking() networkingv1.NetworkingV1Interface {
	ret := _m.Called()

	var r0 networkingv1.NetworkingV1Interface
	if rf, ok := ret.Get(0).(func() networkingv1.NetworkingV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(networkingv1.NetworkingV1Interface)
		}
	}

	return r0
}

// NetworkingV1 provides a mock function with given fields:
func (_m *Interface) NetworkingV1() networkingv1.NetworkingV1Interface {
	ret := _m.Called()

	var r0 networkingv1.NetworkingV1Interface
	if rf, ok := ret.Get(0).(func() networkingv1.NetworkingV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(networkingv1.NetworkingV1Interface)
		}
	}

	return r0
}

// Policy provides a mock function with given fields:
func (_m *Interface) Policy() policyv1beta1.PolicyV1beta1Interface {
	ret := _m.Called()

	var r0 policyv1beta1.PolicyV1beta1Interface
	if rf, ok := ret.Get(0).(func() policyv1beta1.PolicyV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(policyv1beta1.PolicyV1beta1Interface)
		}
	}

	return r0
}

// PolicyV1beta1 provides a mock function with given fields:
func (_m *Interface) PolicyV1beta1() policyv1beta1.PolicyV1beta1Interface {
	ret := _m.Called()

	var r0 policyv1beta1.PolicyV1beta1Interface
	if rf, ok := ret.Get(0).(func() policyv1beta1.PolicyV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(policyv1beta1.PolicyV1beta1Interface)
		}
	}

	return r0
}

// Rbac provides a mock function with given fields:
func (_m *Interface) Rbac() rbacv1.RbacV1Interface {
	ret := _m.Called()

	var r0 rbacv1.RbacV1Interface
	if rf, ok := ret.Get(0).(func() rbacv1.RbacV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rbacv1.RbacV1Interface)
		}
	}

	return r0
}

// RbacV1 provides a mock function with given fields:
func (_m *Interface) RbacV1() rbacv1.RbacV1Interface {
	ret := _m.Called()

	var r0 rbacv1.RbacV1Interface
	if rf, ok := ret.Get(0).(func() rbacv1.RbacV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rbacv1.RbacV1Interface)
		}
	}

	return r0
}

// RbacV1alpha1 provides a mock function with given fields:
func (_m *Interface) RbacV1alpha1() rbacv1alpha1.RbacV1alpha1Interface {
	ret := _m.Called()

	var r0 rbacv1alpha1.RbacV1alpha1Interface
	if rf, ok := ret.Get(0).(func() rbacv1alpha1.RbacV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rbacv1alpha1.RbacV1alpha1Interface)
		}
	}

	return r0
}

// RbacV1beta1 provides a mock function with given fields:
func (_m *Interface) RbacV1beta1() rbacv1beta1.RbacV1beta1Interface {
	ret := _m.Called()

	var r0 rbacv1beta1.RbacV1beta1Interface
	if rf, ok := ret.Get(0).(func() rbacv1beta1.RbacV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rbacv1beta1.RbacV1beta1Interface)
		}
	}

	return r0
}

// Scheduling provides a mock function with given fields:
func (_m *Interface) Scheduling() schedulingv1beta1.SchedulingV1beta1Interface {
	ret := _m.Called()

	var r0 schedulingv1beta1.SchedulingV1beta1Interface
	if rf, ok := ret.Get(0).(func() schedulingv1beta1.SchedulingV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schedulingv1beta1.SchedulingV1beta1Interface)
		}
	}

	return r0
}

// SchedulingV1alpha1 provides a mock function with given fields:
func (_m *Interface) SchedulingV1alpha1() schedulingv1alpha1.SchedulingV1alpha1Interface {
	ret := _m.Called()

	var r0 schedulingv1alpha1.SchedulingV1alpha1Interface
	if rf, ok := ret.Get(0).(func() schedulingv1alpha1.SchedulingV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schedulingv1alpha1.SchedulingV1alpha1Interface)
		}
	}

	return r0
}

// SchedulingV1beta1 provides a mock function with given fields:
func (_m *Interface) SchedulingV1beta1() schedulingv1beta1.SchedulingV1beta1Interface {
	ret := _m.Called()

	var r0 schedulingv1beta1.SchedulingV1beta1Interface
	if rf, ok := ret.Get(0).(func() schedulingv1beta1.SchedulingV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(schedulingv1beta1.SchedulingV1beta1Interface)
		}
	}

	return r0
}

// Settings provides a mock function with given fields:
func (_m *Interface) Settings() settingsv1alpha1.SettingsV1alpha1Interface {
	ret := _m.Called()

	var r0 settingsv1alpha1.SettingsV1alpha1Interface
	if rf, ok := ret.Get(0).(func() settingsv1alpha1.SettingsV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(settingsv1alpha1.SettingsV1alpha1Interface)
		}
	}

	return r0
}

// SettingsV1alpha1 provides a mock function with given fields:
func (_m *Interface) SettingsV1alpha1() settingsv1alpha1.SettingsV1alpha1Interface {
	ret := _m.Called()

	var r0 settingsv1alpha1.SettingsV1alpha1Interface
	if rf, ok := ret.Get(0).(func() settingsv1alpha1.SettingsV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(settingsv1alpha1.SettingsV1alpha1Interface)
		}
	}

	return r0
}

// Storage provides a mock function with given fields:
func (_m *Interface) Storage() storagev1.StorageV1Interface {
	ret := _m.Called()

	var r0 storagev1.StorageV1Interface
	if rf, ok := ret.Get(0).(func() storagev1.StorageV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storagev1.StorageV1Interface)
		}
	}

	return r0
}

// StorageV1 provides a mock function with given fields:
func (_m *Interface) StorageV1() storagev1.StorageV1Interface {
	ret := _m.Called()

	var r0 storagev1.StorageV1Interface
	if rf, ok := ret.Get(0).(func() storagev1.StorageV1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storagev1.StorageV1Interface)
		}
	}

	return r0
}

// StorageV1alpha1 provides a mock function with given fields:
func (_m *Interface) StorageV1alpha1() storagev1alpha1.StorageV1alpha1Interface {
	ret := _m.Called()

	var r0 storagev1alpha1.StorageV1alpha1Interface
	if rf, ok := ret.Get(0).(func() storagev1alpha1.StorageV1alpha1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storagev1alpha1.StorageV1alpha1Interface)
		}
	}

	return r0
}

// StorageV1beta1 provides a mock function with given fields:
func (_m *Interface) StorageV1beta1() storagev1beta1.StorageV1beta1Interface {
	ret := _m.Called()

	var r0 storagev1beta1.StorageV1beta1Interface
	if rf, ok := ret.Get(0).(func() storagev1beta1.StorageV1beta1Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storagev1beta1.StorageV1beta1Interface)
		}
	}

	return r0
}
