package proxy_syncer

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kgateway-dev/kgateway/v2/internal/kgateway/ir"
	"github.com/kgateway-dev/kgateway/v2/internal/kgateway/wellknown"
)

func TestPolicyStatus(t *testing.T) {
	connPolGK := schema.GroupKind{
		Group: "test",
		Kind:  "ConnectionPolicy",
	}
	tlsPolicyAtt := ir.PolicyAtt{
		GroupKind: wellknown.BackendTLSPolicyGVK.GroupKind(),
		PolicyRef: &ir.PolicyRef{
			Group: wellknown.BackendTLSPolicyGVK.Group,
			Kind:  wellknown.BackendTLSPolicyKind,
			Name:  "tls-policy",
		},
		Errors: []error{
			errors.New("error 1"),
		},
	}
	backends := []ir.BackendObjectIR{
		ir.BackendObjectIR{
			ObjectSource: ir.ObjectSource{
				Group:     "",
				Kind:      "Service",
				Namespace: "default",
				Name:      "reviews",
			},
			AttachedPolicies: ir.AttachedPolicies{
				Policies: map[schema.GroupKind][]ir.PolicyAtt{
					wellknown.BackendTLSPolicyGVK.GroupKind(): []ir.PolicyAtt{
						tlsPolicyAtt,
					},
					connPolGK: []ir.PolicyAtt{
						ir.PolicyAtt{
							GroupKind: connPolGK,
							PolicyRef: &ir.PolicyRef{
								Group: connPolGK.Group,
								Kind:  connPolGK.Kind,
								Name:  "conn-policy",
							},
							Errors: []error{},
						},
					},
				},
			},
		},
		ir.BackendObjectIR{
			ObjectSource: ir.ObjectSource{
				Group:     "",
				Kind:      "Service",
				Namespace: "default",
				Name:      "ratings",
			},
			AttachedPolicies: ir.AttachedPolicies{
				Policies: map[schema.GroupKind][]ir.PolicyAtt{
					wellknown.BackendTLSPolicyGVK.GroupKind(): []ir.PolicyAtt{
						tlsPolicyAtt,
					},
					connPolGK: []ir.PolicyAtt{
						ir.PolicyAtt{
							GroupKind: connPolGK,
							PolicyRef: &ir.PolicyRef{
								Group: connPolGK.Group,
								Kind:  connPolGK.Kind,
								Name:  "conn-policy-2",
							},
							Errors: []error{
								errors.New("error 2"),
							},
						},
					},
				},
			},
		},
	}
	seenPolsByGk := generateGkPolicyReport(backends)
	if seenPolsByGk == nil {
		t.Fatalf("seen pols gk is nil")
	}
	t.FailNow()
}

func TestIsGatewayStatusEqual(t *testing.T) {
	addrType := gwv1.HostnameAddressType

	status1 := &gwv1.GatewayStatus{
		Addresses: []gwv1.GatewayStatusAddress{
			{
				Type:  &addrType,
				Value: "address1",
			},
		},
	}
	// same as status1
	status2 := &gwv1.GatewayStatus{
		Addresses: []gwv1.GatewayStatusAddress{
			{
				Type:  &addrType,
				Value: "address1",
			},
		},
	}
	// different from status1
	status3 := &gwv1.GatewayStatus{
		Addresses: []gwv1.GatewayStatusAddress{
			{
				Type:  &addrType,
				Value: "address2",
			},
		},
	}

	tests := []struct {
		name string
		objA *gwv1.GatewayStatus
		objB *gwv1.GatewayStatus
		want bool
	}{
		{"EqualStatus", status1, status2, true},
		{"DifferentStatus", status1, status3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGatewayStatusEqual(tt.objA, tt.objB); got != tt.want {
				t.Errorf("isGatewayStatusEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRouteStatusEqual(t *testing.T) {
	status1 := &gwv1.RouteStatus{
		Parents: []gwv1.RouteParentStatus{
			{
				ParentRef: gwv1.ParentReference{
					Group:     ptr.To[gwv1.Group](gwv1.Group(wellknown.GatewayGroup)),
					Kind:      ptr.To[gwv1.Kind](gwv1.Kind(wellknown.HTTPRouteKind)),
					Name:      "parent",
					Namespace: ptr.To[gwv1.Namespace](gwv1.Namespace("default")),
				},
			},
			{
				ParentRef: gwv1.ParentReference{
					Group:     ptr.To[gwv1.Group](gwv1.Group(wellknown.GatewayGroup)),
					Kind:      ptr.To[gwv1.Kind](gwv1.Kind(wellknown.TCPRouteKind)),
					Name:      "parent",
					Namespace: ptr.To[gwv1.Namespace](gwv1.Namespace("default")),
				},
			},
		},
	}
	// Same as status1
	status2 := &gwv1.RouteStatus{
		Parents: []gwv1.RouteParentStatus{
			{
				ParentRef: gwv1.ParentReference{
					Group:     ptr.To[gwv1.Group](gwv1.Group(wellknown.GatewayGroup)),
					Kind:      ptr.To[gwv1.Kind](gwv1.Kind(wellknown.HTTPRouteKind)),
					Name:      "parent",
					Namespace: ptr.To[gwv1.Namespace](gwv1.Namespace("default")),
				},
			},
			{
				ParentRef: gwv1.ParentReference{
					Group:     ptr.To[gwv1.Group](gwv1.Group(wellknown.GatewayGroup)),
					Kind:      ptr.To[gwv1.Kind](gwv1.Kind(wellknown.TCPRouteKind)),
					Name:      "parent",
					Namespace: ptr.To[gwv1.Namespace](gwv1.Namespace("default")),
				},
			},
		},
	}
	// Different from status1
	status3 := &gwv1.RouteStatus{
		Parents: []gwv1.RouteParentStatus{
			{
				ParentRef: gwv1.ParentReference{
					Group:     ptr.To[gwv1.Group](gwv1.Group(wellknown.GatewayGroup)),
					Kind:      ptr.To[gwv1.Kind](gwv1.Kind(wellknown.HTTPRouteKind)),
					Name:      "parent",
					Namespace: ptr.To[gwv1.Namespace](gwv1.Namespace("my-other-ns")),
				},
			},
			{
				ParentRef: gwv1.ParentReference{
					Group:     ptr.To[gwv1.Group](gwv1.Group(wellknown.GatewayGroup)),
					Kind:      ptr.To[gwv1.Kind](gwv1.Kind(wellknown.TCPRouteKind)),
					Name:      "parent",
					Namespace: ptr.To[gwv1.Namespace](gwv1.Namespace("my-other-ns")),
				},
			},
		},
	}

	tests := []struct {
		name string
		objA *gwv1.RouteStatus
		objB *gwv1.RouteStatus
		want bool
	}{
		{"EqualStatus", status1, status2, true},
		{"DifferentStatus", status1, status3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRouteStatusEqual(tt.objA, tt.objB); got != tt.want {
				t.Errorf("isRouteStatusEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
