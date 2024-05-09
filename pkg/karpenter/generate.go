package karpenter

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"

	"k8s.io/cli-runtime/pkg/printers"
	sigkarpenter "sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

func Generate(nodeGroups *[]types.Nodegroup) error {
	nodePools := make([]sigkarpenter.NodePool, 0, len(*nodeGroups))
	nodeClasses := make([]awskarpenter.EC2NodeClass, 0, len(*nodeGroups))

	for _, ng := range *nodeGroups {
		nodegroup, err := NewNodeGroup(&ng)
		if err != nil {
			return err
		}

		ec2Class, err := nodegroup.GetEC2NodeClass()
		if err != nil {
			return err
		}
		nodeClasses = append(nodeClasses, ec2Class)

		nodePool, err := nodegroup.GetNodePool()
		if err != nil {
			return err
		}
		nodePools = append(nodePools, nodePool)
	}

	return Print(nodePools, nodeClasses)
}

func Print(nodePools []sigkarpenter.NodePool, nodeClasses []awskarpenter.EC2NodeClass) error {
	y := printers.YAMLPrinter{}

	for idx := range len(nodePools) {
		err := y.PrintObj(&nodePools[idx], os.Stdout)
		if err != nil {
			return err
		}
		err = y.PrintObj(&nodeClasses[idx], os.Stdout)
		if err != nil {
			return err
		}
	}
	return nil
}
