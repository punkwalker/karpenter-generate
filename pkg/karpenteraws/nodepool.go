package karpenteraws

import (
	"context"
	"strconv"
	"strings"
	"time"

	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sigkarpenter "sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

var (
	NodePoolTypeMeta = metav1.TypeMeta{
		Kind:       "NodePool",
		APIVersion: sigkarpenter.SchemeGroupVersion.Identifier(),
	}
)

func (n *NodeGroup) GetNodePool() (sigkarpenter.NodePool, error) {
	np := sigkarpenter.NodePool{
		TypeMeta:   NodePoolTypeMeta,
		ObjectMeta: n.NodePoolObjectMeta(),
		Spec:       n.NodePoolSpec(),
	}

	if err := np.Validate(context.TODO()); err != nil {
		return sigkarpenter.NodePool{}, err
	}
	return np, nil
}

func (n NodeGroup) NodePoolObjectMeta() metav1.ObjectMeta {
	min := i32Toa(*n.ScalingConfig.MinSize)
	max := i32Toa(*n.ScalingConfig.MaxSize)
	desired := i32Toa(*n.ScalingConfig.DesiredSize)
	nodePoolAnnotations := map[string]string{
		"generated-by":                 "karpenter-migrate",
		"migrate.karpenter.io/min":     min,
		"migrate.karpenter.io/max":     max,
		"migrate.karpenter.io/desired": desired,
	}

	return metav1.ObjectMeta{
		Name:        n.Name(),
		Annotations: nodePoolAnnotations,
	}
}

func (n NodeGroup) NodePoolSpec() sigkarpenter.NodePoolSpec {
	return sigkarpenter.NodePoolSpec{
		Template: n.NodeClaimTemplate(),
		Disruption: sigkarpenter.Disruption{
			ConsolidateAfter: &sigkarpenter.NillableDuration{
				Duration: lo.ToPtr(30 * time.Second),
			},
			ConsolidationPolicy: sigkarpenter.ConsolidationPolicyWhenEmpty,
		},
		Limits: sigkarpenter.Limits{
			corev1.ResourceCPU: resource.MustParse("1000"),
		},
	}
}

func (n NodeGroup) NodeClaimTemplate() sigkarpenter.NodeClaimTemplate {
	return sigkarpenter.NodeClaimTemplate{
		ObjectMeta: n.NodeClaimObjectMeta(),
		Spec:       n.NodeClaimSpec(),
	}
}

func (n NodeGroup) NodeClaimObjectMeta() sigkarpenter.ObjectMeta {
	return sigkarpenter.ObjectMeta{
		Labels: n.Labels,
	}
}

func (n NodeGroup) NodeClaimSpec() sigkarpenter.NodeClaimSpec {
	return sigkarpenter.NodeClaimSpec{
		NodeClassRef: &sigkarpenter.NodeClassReference{
			Kind:       "EC2NodeClass",
			APIVersion: awskarpenter.SchemeGroupVersion.Identifier(),
			Name:       n.Name(),
		},
		Taints:       n.K8sTaints(),
		Requirements: n.NodeSelectorRequirements(),
	}
}

func (n NodeGroup) NodeSelectorRequirements() []sigkarpenter.NodeSelectorRequirementWithMinValues {
	keys := []string{
		"karpenter.sh/capacity-type",
		"kubernetes.io/arch",
		"node.kubernetes.io/instance-type",
	}
	reqs := []sigkarpenter.NodeSelectorRequirementWithMinValues{}

	for _, key := range keys {
		req := sigkarpenter.NodeSelectorRequirementWithMinValues{}
		switch key {
		case "karpenter.sh/capacity-type":
			req = sigkarpenter.NodeSelectorRequirementWithMinValues{
				NodeSelectorRequirement: corev1.NodeSelectorRequirement{
					Key:      key,
					Operator: "In",
					Values:   n.CapacityTypes(),
				},
			}
		case "kubernetes.io/arch":
			arch := []string{"amd64"}
			if strings.Contains(string(n.AmiType), "ARM") {
				arch = []string{"arm64"}
			}

			req = sigkarpenter.NodeSelectorRequirementWithMinValues{
				NodeSelectorRequirement: corev1.NodeSelectorRequirement{
					Key:      key,
					Operator: "In",
					Values:   arch,
				},
			}
		case "node.kubernetes.io/instance-type":
			req = sigkarpenter.NodeSelectorRequirementWithMinValues{
				NodeSelectorRequirement: corev1.NodeSelectorRequirement{
					Key:      key,
					Operator: "In",
					Values:   n.InstanceTypes,
				},
			}
		}
		reqs = append(reqs, req)
	}
	return reqs
}

func (n NodeGroup) CapacityTypes() []string {
	switch n.CapacityType {
	case ekstypes.CapacityTypesSpot:
		return []string{"spot"}
	default:
		return []string{"on-demand"}
	}
}

func (n NodeGroup) K8sTaints() []corev1.Taint {
	var taints []corev1.Taint
	for _, t := range n.Taints {
		taint := corev1.Taint{
			Key:   *t.Key,
			Value: *t.Value,
		}

		switch t.Effect {
		case ekstypes.TaintEffectNoSchedule:
			taint.Effect = corev1.TaintEffectNoSchedule
		case ekstypes.TaintEffectNoExecute:
			taint.Effect = corev1.TaintEffectNoExecute
		case ekstypes.TaintEffectPreferNoSchedule:
			taint.Effect = corev1.TaintEffectPreferNoSchedule
		}

		taints = append(taints, taint)
	}
	return taints
}

func i32Toa(i int32) string {
	return strconv.Itoa(int(i))
}
