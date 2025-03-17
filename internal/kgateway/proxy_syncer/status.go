package proxy_syncer

import (
	"maps"
	"slices"

	plug "github.com/kgateway-dev/kgateway/v2/internal/kgateway/extensions2/plugin"
	"github.com/kgateway-dev/kgateway/v2/internal/kgateway/ir"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type attachmentReport struct {
	Errors   []error
	Ancestor ir.ObjectSource
}
type policyObjsWithReports map[ir.PolicyRef][]attachmentReport

type GKPolicyReport struct {
	SeenPolicies map[schema.GroupKind]plug.PolicyReport
}

func (r GKPolicyReport) ResourceName() string {
	return "report"
}

func comparePolicyErrors(i, j plug.PolicyErrors) bool {
	if !slices.Equal(i.Errors, j.Errors) {
		return false
	}
	return true
}
func comparePolicyWithAncestorReports(x, y plug.PolicyWithAncestorReports) bool {
	if !maps.EqualFunc(x.AncestorReports, y.AncestorReports, comparePolicyErrors) {
		return false
	}
	return true
}

func (r GKPolicyReport) Equals(in GKPolicyReport) bool {
	for gk, reports := range r.SeenPolicies {
		inreports, ok := in.SeenPolicies[gk]
		if !ok {
			return false
		}
		if !maps.EqualFunc(reports, inreports, comparePolicyWithAncestorReports) {
			return false
		}
	}
	return true
}

func generateGkPolicyReport(backends []ir.BackendObjectIR) *GKPolicyReport {
	seenPolicyResources := policyObjsWithReports{}
	for _, backendObj := range backends {
		for _, polAtts := range backendObj.AttachedPolicies.Policies {
			for _, polAtt := range polAtts {
				if polAtt.PolicyRef == nil {
					// the policyRef may be nil in the case of virtual plugins (e.g. istio settings)
					// since there's no real policy object, we don't need to generate status for it
					continue
				}
				ar := attachmentReport{
					Ancestor: backendObj.ObjectSource,
					Errors:   polAtt.Errors,
				}
				reports := seenPolicyResources[*polAtt.PolicyRef]
				reports = append(reports, ar)
				seenPolicyResources[*polAtt.PolicyRef] = reports
			}
		}
	}
	// seenPolsByGk := map[schema.GroupKind][]policyWithAncestorReports{}
	policiesToAncestorReports := plug.PolicyReport{}
	for policyRef, reports := range seenPolicyResources {
		ancestorReports := map[ir.ObjectSource]plug.PolicyErrors{}
		for _, rpt := range reports {
			ancestorReports[rpt.Ancestor] = plug.PolicyErrors{Errors: rpt.Errors}
		}
		policiesToAncestorReports[policyRef] = plug.PolicyWithAncestorReports{
			AncestorReports: ancestorReports,
		}
	}

	seenPolsByGk := map[schema.GroupKind]plug.PolicyReport{}
	for policyRef, reports := range policiesToAncestorReports {
		gk := schema.GroupKind{
			Group: policyRef.Group,
			Kind:  policyRef.Kind,
		}
		gkPolsMap := seenPolsByGk[gk]
		if gkPolsMap == nil {
			gkPolsMap = map[ir.PolicyRef]plug.PolicyWithAncestorReports{}
		}
		gkPolsMap[policyRef] = reports
		seenPolsByGk[gk] = gkPolsMap
	}
	return &GKPolicyReport{
		SeenPolicies: seenPolsByGk,
	}
}
