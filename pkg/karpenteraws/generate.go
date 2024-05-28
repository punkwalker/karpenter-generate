package karpenteraws

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	"github.com/punkwalker/karpenter-generate/pkg/aws"
	sigkarpenter "sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

type NodeGroup struct {
	*ekstypes.Nodegroup
	LT       *ec2types.ResponseLaunchTemplateData // LT Generated by MNG and used by ASG (needed for MetadataOptions)
	CustomLT *ec2types.ResponseLaunchTemplateData // Custom LT provided to MNG
}

func Generate(nodeGroups *[]ekstypes.Nodegroup) ([]sigkarpenter.NodePool, []awskarpenter.EC2NodeClass, error) {
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

func NewNodeGroup(ng ekstypes.Nodegroup) (*NodeGroup, error) {

	newNodegroup := NodeGroup{
		Nodegroup: &ng,
	}

	ClusterTag = map[string]string{
		ClusterTagKey + *ng.ClusterName: "owned",
	}

	ec2Client := aws.NewEC2Client()
	asgClient := aws.NewAutoscalingClient()

	if ng.LaunchTemplate != nil {
		customLT, err := ec2Client.DescribeLaunchTemplateVersions(
			*ng.LaunchTemplate.Id,
			*ng.LaunchTemplate.Version)
		if err != nil {
			return nil, aws.FormatErrorAsMessageOnly(err)
		}
		newNodegroup.CustomLT = customLT[0].LaunchTemplateData
	}

	asg, err := asgClient.DescribeAutoScalingGroups(*ng.Resources.AutoScalingGroups[0].Name)
	if err != nil {
		return nil, aws.FormatErrorAsMessageOnly(err)
	}

	lt, err := ec2Client.DescribeLaunchTemplateVersions(
		*asg[0].MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId,
		*asg[0].MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version)
	if err != nil {
		return nil, aws.FormatErrorAsMessageOnly(err)
	}

	newNodegroup.LT = lt[0].LaunchTemplateData

	return &newNodegroup, nil
}
