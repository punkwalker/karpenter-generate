package karpenteraws

import (
	"reflect"
	"testing"

	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sigkarpenter "sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

func TestNodeGroup_NodePoolObjectMeta(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected metav1.ObjectMeta
	}{
		{
			name: "Valid scaling config",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					NodegroupName: lo.ToPtr("my-node-group"),
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						MinSize:     lo.ToPtr(int32(1)),
						MaxSize:     lo.ToPtr(int32(10)),
						DesiredSize: lo.ToPtr(int32(5)),
					},
				},
			},
			expected: metav1.ObjectMeta{
				Name: "my-node-group",
				Annotations: map[string]string{
					"generated-by":                 "karpenter-migrate",
					"migrate.karpenter.io/min":     "1",
					"migrate.karpenter.io/max":     "10",
					"migrate.karpenter.io/desired": "5",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.NodePoolObjectMeta()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.NodePoolObjectMeta() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_NodeClaimObjectMeta(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected sigkarpenter.ObjectMeta
	}{
		{
			name: "Nodegroup with Labels",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					Labels: map[string]string{
						"label1": "val1",
						"label2": "val2",
					},
				},
			},
			expected: sigkarpenter.ObjectMeta{
				Labels: map[string]string{
					"label1": "val1",
					"label2": "val2",
				},
			},
		},
		{
			name: "Nodegroup Without Labels",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					Labels: nil,
				},
			},
			expected: sigkarpenter.ObjectMeta{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.NodeClaimObjectMeta()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.NodeClaimObjectMeta() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_NodeSelectorRequirements(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected []sigkarpenter.NodeSelectorRequirementWithMinValues
	}{
		{
			name: "NodeGroup with OnDemand capacity and Multiple instance types and X86_64 AMI",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					CapacityType:  ekstypes.CapacityTypesOnDemand,
					InstanceTypes: []string{"t2.micro", "t3.micro"},
					AmiType:       ekstypes.AMITypesAl2X8664,
				},
			},
			expected: []sigkarpenter.NodeSelectorRequirementWithMinValues{
				{
					NodeSelectorRequirement: corev1.NodeSelectorRequirement{
						Key:      "karpenter.sh/capacity-type",
						Operator: "In",
						Values:   []string{"on-demand"},
					},
				},
				{
					NodeSelectorRequirement: corev1.NodeSelectorRequirement{
						Key:      "kubernetes.io/arch",
						Operator: "In",
						Values:   []string{"amd64"},
					},
				},
				{
					NodeSelectorRequirement: corev1.NodeSelectorRequirement{
						Key:      "node.kubernetes.io/instance-type",
						Operator: "In",
						Values:   []string{"t2.micro", "t3.micro"},
					},
				},
			},
		},
		{
			name: "NodeGroup with Spot capacity and ARM64 AMI",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					CapacityType:  ekstypes.CapacityTypesSpot,
					InstanceTypes: []string{"t3.micro"},
					AmiType:       ekstypes.AMITypesBottlerocketArm64,
				},
			},
			expected: []sigkarpenter.NodeSelectorRequirementWithMinValues{
				{
					NodeSelectorRequirement: corev1.NodeSelectorRequirement{
						Key:      "karpenter.sh/capacity-type",
						Operator: "In",
						Values:   []string{"spot"},
					},
				},
				{
					NodeSelectorRequirement: corev1.NodeSelectorRequirement{
						Key:      "kubernetes.io/arch",
						Operator: "In",
						Values:   []string{"arm64"},
					},
				},
				{
					NodeSelectorRequirement: corev1.NodeSelectorRequirement{
						Key:      "node.kubernetes.io/instance-type",
						Operator: "In",
						Values:   []string{"t3.micro"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.NodeSelectorRequirements()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.NodeSelectorRequirements() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_CapacityTypes(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected []string
	}{
		{
			name: "Spot capacity type",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					CapacityType: ekstypes.CapacityTypesSpot,
				},
			},
			expected: []string{"spot"},
		},
		{
			name: "On-demand capacity type",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					CapacityType: ekstypes.CapacityTypesOnDemand,
				},
			},
			expected: []string{"on-demand"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.CapacityTypes()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.CapacityTypes() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_K8sTaints(t *testing.T) {

	var emptyTaints []corev1.Taint
	tests := []struct {
		name     string
		n        NodeGroup
		expected []corev1.Taint
	}{
		{
			name: "Nodegroup with taints",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					Taints: []ekstypes.Taint{
						{
							Key:    lo.ToPtr("key1"),
							Value:  lo.ToPtr("value1"),
							Effect: ekstypes.TaintEffectNoSchedule,
						},
						{
							Key:    lo.ToPtr("key2"),
							Value:  lo.ToPtr("value2"),
							Effect: ekstypes.TaintEffectNoExecute,
						},
						{
							Key:    lo.ToPtr("key3"),
							Value:  lo.ToPtr("value3"),
							Effect: ekstypes.TaintEffectPreferNoSchedule,
						},
					},
				},
			},
			expected: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: corev1.TaintEffectNoExecute,
				},
				{
					Key:    "key3",
					Value:  "value3",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			},
		},
		{
			name: "Nodegroup without taints",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{},
			},
			expected: emptyTaints,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.K8sTaints()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.K8sTaints() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
