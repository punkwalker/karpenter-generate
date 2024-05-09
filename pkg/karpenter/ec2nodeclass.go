package karpenter

import (
	"strings"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/punkwalker/karpenter-generate/pkg/aws"
)

type NodeGroup struct {
	*ekstypes.Nodegroup
	LtData *ec2types.ResponseLaunchTemplateData
}

var (
	NodeClassTypeMeta = metav1.TypeMeta{
		Kind:       "EC2NodeClass",
		APIVersion: awskarpenter.SchemeGroupVersion.Identifier(),
	}
)

func NewNodeGroup(ng ekstypes.Nodegroup) (*NodeGroup, error) {

	ec2Client := aws.NewEC2Client()
	asgClient := aws.NewAutoscalingClient()

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

	return &NodeGroup{
		Nodegroup: &ng,
		LtData:    lt[0].LaunchTemplateData,
	}, nil
}

func (n *NodeGroup) GetEC2NodeClass() (awskarpenter.EC2NodeClass, error) {
	return awskarpenter.EC2NodeClass{
		TypeMeta:   NodeClassTypeMeta,
		ObjectMeta: n.NodeClassObjectMeta(),
		Spec:       n.NodeClassSpec(),
	}, nil
}

func (n NodeGroup) Name() string {
	return *n.NodegroupName
}

func (n NodeGroup) AmiID() string {
	return *n.LtData.ImageId
}

func (n NodeGroup) NodeClassObjectMeta() metav1.ObjectMeta {
	nodeClassannotations := map[string]string{
		"generated-by": "karpenter-migrate",
	}

	return metav1.ObjectMeta{
		Name:        n.Name(),
		Annotations: nodeClassannotations,
	}
}

func (n NodeGroup) NodeClassSpec() awskarpenter.EC2NodeClassSpec {
	amiFamily := n.AMIFamily()
	if amiFamily != &awskarpenter.AMIFamilyCustom {
		return awskarpenter.EC2NodeClassSpec{
			AMIFamily:                  amiFamily,
			Role:                       n.Role(),
			SubnetSelectorTerms:        n.SubnetSelectorTerms(),
			SecurityGroupSelectorTerms: n.SecurityGroupSelectorTerms(),
			Tags:                       n.FilteredTags(),
		}
	}

	return awskarpenter.EC2NodeClassSpec{
		AMIFamily:        amiFamily,
		AMISelectorTerms: n.AMISelectorTerms(),
	}

}

func (n NodeGroup) FilteredTags() map[string]string {
	filteredTags := map[string]string{}
	for key, val := range n.Tags {
		if !strings.HasPrefix(key, "aws:") {
			filteredTags[key] = val
		}
	}
	return filteredTags
}

func (n NodeGroup) AMIFamily() *string {
	switch n.AmiType {
	case ekstypes.AMITypesAl2X8664, ekstypes.AMITypesAl2X8664Gpu, ekstypes.AMITypesAl2Arm64:
		return &awskarpenter.AMIFamilyAL2
	case ekstypes.AMITypesBottlerocketX8664, ekstypes.AMITypesBottlerocketArm64, ekstypes.AMITypesBottlerocketArm64Nvidia, ekstypes.AMITypesBottlerocketX8664Nvidia:
		return &awskarpenter.AMIFamilyBottlerocket
	case ekstypes.AMITypesWindowsFull2019X8664, ekstypes.AMITypesWindowsCore2019X8664:
		return &awskarpenter.AMIFamilyWindows2019
	case ekstypes.AMITypesWindowsFull2022X8664, ekstypes.AMITypesWindowsCore2022X8664:
		return &awskarpenter.AMIFamilyWindows2022
	case ekstypes.AMITypesAl2023X8664Standard, ekstypes.AMITypesAl2023Arm64Standard:
		return &awskarpenter.AMIFamilyAL2023
	default:
		return &awskarpenter.AMIFamilyCustom
	}
}

func (n NodeGroup) AMISelectorTerms() []awskarpenter.AMISelectorTerm {
	amiTerms := []awskarpenter.AMISelectorTerm{}
	amiTerms = append(amiTerms, awskarpenter.AMISelectorTerm{
		ID: n.AmiID(),
	})
	return amiTerms
}

func (n NodeGroup) Role() string {
	roleName := *n.NodeRole
	lastIdx := strings.LastIndex(*n.NodeRole, "/")
	roleName = roleName[lastIdx+1:]
	return roleName
}

func (n NodeGroup) SubnetSelectorTerms() []awskarpenter.SubnetSelectorTerm {
	subnetSlice := []awskarpenter.SubnetSelectorTerm{}
	for _, subnet := range n.Subnets {
		subnetSlice = append(subnetSlice, awskarpenter.SubnetSelectorTerm{
			ID: subnet,
		})
	}
	return subnetSlice
}

func (n NodeGroup) SecurityGroupSelectorTerms() []awskarpenter.SecurityGroupSelectorTerm {
	sgTerms := []awskarpenter.SecurityGroupSelectorTerm{}
	ltSG := n.LtData.SecurityGroupIds

	// For LTs created by simple MNG
	for _, intf := range n.LtData.NetworkInterfaces {
		for _, sg := range intf.Groups {
			sgTerms = append(sgTerms, awskarpenter.SecurityGroupSelectorTerm{
				ID: sg,
			})
		}
	}

	// For customLaunchTemplates
	for _, sg := range ltSG {
		sgTerms = append(sgTerms, awskarpenter.SecurityGroupSelectorTerm{
			ID: sg,
		})
	}
	return sgTerms
}
