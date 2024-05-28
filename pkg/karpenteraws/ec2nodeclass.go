package karpenteraws

import (
	"context"
	"encoding/base64"
	"strings"

	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	awskarpenterprovider "github.com/aws/karpenter-provider-aws/pkg/providers/amifamily"
	"github.com/samber/lo"
	k8sapiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GiB                              int64  = 1024 * 1024 * 1024
	ClusterTagKey                    string = "kubernetes.io/cluster/"
	ALAndBottleRocketDefaultDiskSize        = 20
	WindowsDefaultDiskSize                  = 50
)

var (
	NodeClassTypeMeta = metav1.TypeMeta{
		Kind:       "EC2NodeClass",
		APIVersion: awskarpenter.SchemeGroupVersion.Identifier(),
	}
	ClusterTag map[string]string
)

func (n *NodeGroup) GetEC2NodeClass() (awskarpenter.EC2NodeClass, error) {
	nc := awskarpenter.EC2NodeClass{
		TypeMeta:   NodeClassTypeMeta,
		ObjectMeta: n.NodeClassObjectMeta(),
		Spec:       n.NodeClassSpec(),
	}
	if err := nc.Validate(context.TODO()); err != nil {
		return awskarpenter.EC2NodeClass{}, err
	}
	return nc, nil
}

func (n NodeGroup) Name() string {
	return strings.ToLower(*n.NodegroupName)
}

func (n NodeGroup) AmiID() string {
	// TODO: implement preserve AMIId for MNG
	return *n.CustomLT.ImageId
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
	return awskarpenter.EC2NodeClassSpec{
		AMIFamily:                  n.AMIFamily(),
		Role:                       n.Role(),
		AMISelectorTerms:           n.AMISelectorTerms(),
		SubnetSelectorTerms:        n.SubnetSelectorTerms(),
		SecurityGroupSelectorTerms: n.SecurityGroupSelectorTerms(),
		UserData:                   n.UserData(),
		BlockDeviceMappings:        n.BlockDeviceMappings(),
		Tags:                       n.FilteredTags(),
		MetadataOptions:            n.MetadataOptions(),
	}
}

func (n NodeGroup) FilteredTags() map[string]string {
	filteredTags := map[string]string{}
	if n.CustomLT != nil {
		for _, tagspec := range n.CustomLT.TagSpecifications {
			for _, tag := range tagspec.Tags {
				if !strings.HasPrefix(*tag.Key, "aws:") {
					filteredTags[*tag.Key] = *tag.Value
				}
			}
		}
	}

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
		return lo.ToPtr(awskarpenter.AMIFamilyAL2)
	case ekstypes.AMITypesBottlerocketX8664, ekstypes.AMITypesBottlerocketArm64, ekstypes.AMITypesBottlerocketArm64Nvidia, ekstypes.AMITypesBottlerocketX8664Nvidia:
		return lo.ToPtr(awskarpenter.AMIFamilyBottlerocket)
	case ekstypes.AMITypesWindowsFull2019X8664, ekstypes.AMITypesWindowsCore2019X8664:
		return lo.ToPtr(awskarpenter.AMIFamilyWindows2019)
	case ekstypes.AMITypesWindowsFull2022X8664, ekstypes.AMITypesWindowsCore2022X8664:
		return lo.ToPtr(awskarpenter.AMIFamilyWindows2022)
	case ekstypes.AMITypesAl2023X8664Standard, ekstypes.AMITypesAl2023Arm64Standard:
		return lo.ToPtr(awskarpenter.AMIFamilyAL2023)
	default:
		return lo.ToPtr(awskarpenter.AMIFamilyCustom)
	}
}

func (n NodeGroup) IsCustomAMIFamily() bool {
	return n.AmiType == ekstypes.AMITypesCustom
}

func (n NodeGroup) AMISelectorTerms() []awskarpenter.AMISelectorTerm {
	amiTerms := []awskarpenter.AMISelectorTerm{}
	if n.CustomLT != nil {
		if n.CustomLT.ImageId != nil {
			amiTerms = append(amiTerms, awskarpenter.AMISelectorTerm{
				ID: n.AmiID(),
			})
		}
	}
	return amiTerms
}

func (n NodeGroup) Role() string {
	// TODO: Implement override for MNG nodeRole
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

	// For customLaunchTemplates
	if n.CustomLT != nil {
		if n.CustomLT.SecurityGroupIds != nil {
			ltSG := n.CustomLT.SecurityGroupIds
			for _, sg := range ltSG {
				sgTerms = append(sgTerms, awskarpenter.SecurityGroupSelectorTerm{
					ID: sg,
				})
			}
		}
		if n.CustomLT.SecurityGroups != nil {
			ltSG := n.CustomLT.SecurityGroups
			for _, sg := range ltSG {
				sgTerms = append(sgTerms, awskarpenter.SecurityGroupSelectorTerm{
					Name: sg,
				})
			}
		}
		if n.CustomLT.NetworkInterfaces != nil {
			for _, eni := range n.CustomLT.NetworkInterfaces {
				if eni.Groups != nil {
					for _, sg := range eni.Groups {
						sgTerms = append(sgTerms, awskarpenter.SecurityGroupSelectorTerm{
							ID: sg,
						})
					}
				}
			}
		}
	}

	// Use clusterTag as SecurityGroupSelector if no SGs found
	if len(sgTerms) == 0 {
		sgTerms = append(sgTerms, awskarpenter.SecurityGroupSelectorTerm{
			Tags: ClusterTag,
		})
	}
	return sgTerms
}

// Returns UserData for nodegroup if Custom Launch Template is used with MNG
func (n NodeGroup) UserData() *string {
	if n.CustomLT != nil && n.CustomLT.UserData != nil {
		decodedUserData, _ := base64.StdEncoding.DecodeString(*n.CustomLT.UserData)
		return lo.ToPtr(string(decodedUserData))
	}
	return nil
}

// Returns AWS Karpenter BlockDeviceMappings for nodegroup if Custom Launch Template is used with MNG or Custom DiskSize is configured
func (n NodeGroup) BlockDeviceMappings() []*awskarpenter.BlockDeviceMapping {
	mappings := []*awskarpenter.BlockDeviceMapping{}
	var diskSize *k8sapiresource.Quantity

	if n.CustomLT != nil {
		for _, mapping := range n.CustomLT.BlockDeviceMappings {
			if mapping.Ebs != nil {
				bMap := &awskarpenter.BlockDeviceMapping{
					DeviceName: lo.EmptyableToPtr(lo.FromPtr(mapping.DeviceName)),
					EBS: &awskarpenter.BlockDevice{
						VolumeSize:          k8sapiresource.NewQuantity(int64(*mapping.Ebs.VolumeSize)*GiB, k8sapiresource.BinarySI),
						VolumeType:          lo.ToPtr(string(mapping.Ebs.VolumeType)),
						DeleteOnTermination: mapping.Ebs.DeleteOnTermination,
						Encrypted:           mapping.Ebs.Encrypted,
						IOPS:                lo.EmptyableToPtr(int64(lo.FromPtr(mapping.Ebs.Iops))),
						KMSKeyID:            mapping.Ebs.KmsKeyId,
						SnapshotID:          mapping.Ebs.SnapshotId,
						Throughput:          lo.EmptyableToPtr(int64(lo.FromPtr(mapping.Ebs.Throughput))),
					},
				}
				mappings = append(mappings, bMap)
			}
		}
		return mappings
	}

	// Managed Node Group DiskSize - https: //docs.aws.amazon.com/eks/latest/APIReference/API_CreateNodegroup.html#AmazonEKS-CreateNodegroup-request-diskSize
	if *n.DiskSize != ALAndBottleRocketDefaultDiskSize && (n.AMIFamily() == &awskarpenter.AMIFamilyAL2 || n.AMIFamily() == &awskarpenter.AMIFamilyAL2023 || n.AMIFamily() == &awskarpenter.AMIFamilyBottlerocket) {
		diskSize = k8sapiresource.NewQuantity(int64(*n.DiskSize)*GiB, k8sapiresource.BinarySI)
	}
	if *n.DiskSize != WindowsDefaultDiskSize && (n.AMIFamily() == &awskarpenter.AMIFamilyWindows2019 || n.AMIFamily() == &awskarpenter.AMIFamilyWindows2022) {
		diskSize = k8sapiresource.NewQuantity(int64(*n.DiskSize)*GiB, k8sapiresource.BinarySI)
	}

	// Update the diskSize in default mappings if value is Set
	if diskSize != nil {
		amiFamily := awskarpenterprovider.GetAMIFamily(n.AMIFamily(), &awskarpenterprovider.Options{})
		mappings = amiFamily.DefaultBlockDeviceMappings()
		for _, mapping := range mappings {
			mapping.EBS.VolumeSize = k8sapiresource.NewQuantity(int64(*n.DiskSize)*GiB, k8sapiresource.BinarySI)
		}
	}
	return mappings
}

func (n NodeGroup) MetadataOptions() *awskarpenter.MetadataOptions {
	if n.CustomLT != nil && n.CustomLT.MetadataOptions != nil {
		metaOptions := &awskarpenter.MetadataOptions{
			HTTPEndpoint:            lo.EmptyableToPtr(string(lo.FromPtr(&n.CustomLT.MetadataOptions.HttpEndpoint))),
			HTTPPutResponseHopLimit: lo.EmptyableToPtr(int64(lo.FromPtr(n.LT.MetadataOptions.HttpPutResponseHopLimit))),
			HTTPTokens:              lo.EmptyableToPtr(string(lo.FromPtr(&n.CustomLT.MetadataOptions.HttpTokens))),
			HTTPProtocolIPv6:        lo.EmptyableToPtr(string(lo.FromPtr(&n.LT.MetadataOptions.HttpProtocolIpv6))),
		}
		return metaOptions
	}
	return nil
}
