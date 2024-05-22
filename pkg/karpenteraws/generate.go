package karpenteraws

import (
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	sigkarpenter "sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

func Generate(nodeGroups *[]types.Nodegroup) ([]sigkarpenter.NodePool, []awskarpenter.EC2NodeClass, error) {
	nodePools := make([]sigkarpenter.NodePool, 0, len(*nodeGroups))
	nodeClasses := make([]awskarpenter.EC2NodeClass, 0, len(*nodeGroups))

	for _, ng := range *nodeGroups {
		nodegroup, err := NewNodeGroup(ng)
		if err != nil {
			return nil, nil, err
		}

		ec2Class, err := nodegroup.GetEC2NodeClass()
		if err != nil {
			return nil, nil, err
		}
		nodeClasses = append(nodeClasses, ec2Class)

		nodePool, err := nodegroup.GetNodePool()
		if err != nil {
			return nil, nil, err
		}
		nodePools = append(nodePools, nodePool)
	}

	return nodePools, nodeClasses, nil
}
