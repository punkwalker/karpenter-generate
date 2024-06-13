package karpenteraws

import (
	"reflect"
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	awskarpenterprovider "github.com/aws/karpenter-provider-aws/pkg/providers/amifamily"
	"github.com/samber/lo"
	k8sapiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeGroup_Name(t *testing.T) {
	tests := []struct {
		name string
		n    NodeGroup
		want string
	}{
		{
			name: "Uppercase name",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					NodegroupName: lo.ToPtr("MyNodeGroup"),
				},
			},
			want: "mynodegroup",
		},
		{
			name: "Mixed case name",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					NodegroupName: lo.ToPtr("MyNodeGroup123"),
				},
			},
			want: "mynodegroup123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.Name(); got != tt.want {
				t.Errorf("NodeGroup.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_AmiID(t *testing.T) {
	tests := []struct {
		name string
		n    NodeGroup
		want string
	}{
		{
			name: "Valid AMI ID",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					ImageId: lo.ToPtr("ami-0123456789abcdef"),
				},
			},
			want: "ami-0123456789abcdef",
		},
		{
			name: "Nil CustomLT",
			n: NodeGroup{
				CustomLT: nil,
			},
			want: "",
		},
		{
			name: "Nil ImageId",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					ImageId: nil,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.AmiID(); got != tt.want {
				t.Errorf("NodeGroup.AmiID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_NodeClassObjectMeta(t *testing.T) {
	tests := []struct {
		name string
		n    NodeGroup
		want metav1.ObjectMeta
	}{
		{
			name: "Valid NodeGroup",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					NodegroupName: lo.ToPtr("My-Node-Group"),
				},
			},
			want: metav1.ObjectMeta{
				Name: "my-node-group",
				Annotations: map[string]string{
					"generated-by":                          "karpenter-migrate",
					"migrate.karpenter.sh/source-nodegroup": "my-node-group",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.NodeClassObjectMeta(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroup.NodeClassObjectMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_FilteredTags(t *testing.T) {
	tests := []struct {
		name string
		n    NodeGroup
		want map[string]string
	}{
		{
			name: "No custom tags, no AWS tags",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					TagSpecifications: nil,
				},
				Nodegroup: &ekstypes.Nodegroup{
					Tags: map[string]string{
						"env":  "prod",
						"team": "engineering",
					},
				},
			},
			want: map[string]string{
				"env":  "prod",
				"team": "engineering",
			},
		},
		{
			name: "Custom tags, no AWS tags",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					TagSpecifications: []ec2types.LaunchTemplateTagSpecification{
						{
							Tags: []ec2types.Tag{
								{
									Key:   lo.ToPtr("env"),
									Value: lo.ToPtr("prod"),
								},
								{
									Key:   lo.ToPtr("team"),
									Value: lo.ToPtr("engineering"),
								},
								{
									Key:   lo.ToPtr("aws:tag"),
									Value: lo.ToPtr("value"),
								},
							},
						},
					},
				},
				Nodegroup: &ekstypes.Nodegroup{
					Tags: map[string]string{
						"env":  "dev",
						"team": "devops",
					},
				},
			},
			want: map[string]string{
				"env":  "prod",
				"team": "engineering",
			},
		},
		{
			name: "No custom tags, AWS tags",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					TagSpecifications: nil,
				},
				Nodegroup: &ekstypes.Nodegroup{
					Tags: map[string]string{
						"env":        "prod",
						"team":       "engineering",
						"aws:tag":    "value",
						"aws:prefix": "value",
					},
				},
			},
			want: map[string]string{
				"env":  "prod",
				"team": "engineering",
			},
		},
		{
			name: "Custom tags, AWS tags",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					TagSpecifications: []ec2types.LaunchTemplateTagSpecification{
						{
							Tags: []ec2types.Tag{
								{
									Key:   lo.ToPtr("env"),
									Value: lo.ToPtr("prod"),
								},
								{
									Key:   lo.ToPtr("team"),
									Value: lo.ToPtr("engineering"),
								},
								{
									Key:   lo.ToPtr("aws:tag"),
									Value: lo.ToPtr("value"),
								},
							},
						},
					},
				},
				Nodegroup: &ekstypes.Nodegroup{
					Tags: map[string]string{
						"env":        "dev",
						"team":       "devops",
						"aws:tag":    "value",
						"aws:prefix": "value",
					},
				},
			},
			want: map[string]string{
				"env":  "prod",
				"team": "engineering",
			},
		},
		{
			name: "Nil CustomLT",
			n: NodeGroup{
				CustomLT: nil,
				Nodegroup: &ekstypes.Nodegroup{
					Tags: map[string]string{
						"env":  "prod",
						"team": "engineering",
					},
				},
			},
			want: map[string]string{
				"env":  "prod",
				"team": "engineering",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.FilteredTags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroup.FilteredTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_AMIFamily(t *testing.T) {
	tests := []struct {
		name string
		n    NodeGroup
		want string
	}{
		{
			name: "AL2 X86_64",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType: ekstypes.AMITypesAl2X8664,
				},
			},
			want: awskarpenter.AMIFamilyAL2,
		},
		{
			name: "AL2023 X86_64",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType: ekstypes.AMITypesAl2023X8664Standard,
				},
			},
			want: awskarpenter.AMIFamilyAL2023,
		},
		{
			name: "Bottleorcket X86_64",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType: ekstypes.AMITypesBottlerocketX8664,
				},
			},
			want: awskarpenter.AMIFamilyBottlerocket,
		},
		{
			name: "Windows 2019 X86_64",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType: ekstypes.AMITypesWindowsFull2019X8664,
				},
			},
			want: awskarpenter.AMIFamilyWindows2019,
		},
		{
			name: "Windows 2022 X86_64",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType: ekstypes.AMITypesWindowsFull2022X8664,
				},
			},
			want: awskarpenter.AMIFamilyWindows2022,
		},
		{
			name: "Custom AMI",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType: ekstypes.AMITypesCustom,
				},
			},
			want: awskarpenter.AMIFamilyCustom,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := *tt.n.AMIFamily(); got != tt.want {
				t.Errorf("NodeGroup.AMIFamily() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_AMISelectorTerms(t *testing.T) {
	tests := []struct {
		name string
		n    NodeGroup
		want []awskarpenter.AMISelectorTerm
	}{
		{
			name: "Valid AMI ID",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					ImageId: lo.ToPtr("ami-0123456789abcdef"),
				},
			},
			want: []awskarpenter.AMISelectorTerm{
				{
					ID: "ami-0123456789abcdef",
				},
			},
		},
		{
			name: "Nil CustomLT",
			n: NodeGroup{
				CustomLT: nil,
			},
			want: []awskarpenter.AMISelectorTerm{},
		},
		{
			name: "Nil ImageId",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					ImageId: nil,
				},
			},
			want: []awskarpenter.AMISelectorTerm{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.AMISelectorTerms(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroup.AMISelectorTerms() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_Role(t *testing.T) {
	tests := []struct {
		name string
		n    NodeGroup
		want string
	}{
		{
			name: "Valid role ARN",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					NodeRole: lo.ToPtr("arn:aws:iam::123456789012:role/my-node-role"),
				},
			},
			want: "my-node-role",
		},
		{
			name: "Role ARN with multiple /",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					NodeRole: lo.ToPtr("arn:aws:iam::123456789012:role/role-group/my-node-role"),
				},
			},
			want: "my-node-role",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.Role(); got != tt.want {
				t.Errorf("NodeGroup.Role() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_SubnetSelectorTerms(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected []awskarpenter.SubnetSelectorTerm
	}{
		{
			name: "Single subnet",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					Subnets: []string{"subnet-0123456789abcdef"},
				},
			},
			expected: []awskarpenter.SubnetSelectorTerm{
				{
					ID: "subnet-0123456789abcdef",
				},
			},
		},
		{
			name: "Multiple subnets",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					Subnets: []string{"subnet-0123456789abcdef", "subnet-fedcba9876543210"},
				},
			},
			expected: []awskarpenter.SubnetSelectorTerm{
				{
					ID: "subnet-0123456789abcdef",
				},
				{
					ID: "subnet-fedcba9876543210",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.SubnetSelectorTerms()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.SubnetSelectorTerms() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_SecurityGroupSelectorTerms(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected []awskarpenter.SecurityGroupSelectorTerm
	}{
		{
			name: "Security group IDs in CustomLT",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					SecurityGroupIds: []string{"sg-0123456789abcdef", "sg-fedcba9876543210"},
				},
			},
			expected: []awskarpenter.SecurityGroupSelectorTerm{
				{
					ID: "sg-0123456789abcdef",
				},
				{
					ID: "sg-fedcba9876543210",
				},
			},
		},
		{
			name: "Security group names in CustomLT",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					SecurityGroups: []string{"my-security-group", "another-security-group"},
				},
			},
			expected: []awskarpenter.SecurityGroupSelectorTerm{
				{
					Name: "my-security-group",
				},
				{
					Name: "another-security-group",
				},
			},
		},
		{
			name: "Security groups in network interfaces in CustomLT",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					NetworkInterfaces: []ec2types.LaunchTemplateInstanceNetworkInterfaceSpecification{
						{
							Groups: []string{"sg-0123456789abcdef", "sg-fedcba9876543210"},
						},
					},
				},
			},
			expected: []awskarpenter.SecurityGroupSelectorTerm{
				{
					ID: "sg-0123456789abcdef",
				},
				{
					ID: "sg-fedcba9876543210",
				},
			},
		},
		{
			name: "No security groups in CustomLT",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					SecurityGroupIds: nil,
					SecurityGroups:   nil,
					NetworkInterfaces: []ec2types.LaunchTemplateInstanceNetworkInterfaceSpecification{
						{
							Groups: nil,
						},
					},
				},
			},
			expected: []awskarpenter.SecurityGroupSelectorTerm{
				{
					Tags: ClusterTag,
				},
			},
		},
		{
			name: "Nil CustomLT",
			n: NodeGroup{
				CustomLT: nil,
			},
			expected: []awskarpenter.SecurityGroupSelectorTerm{
				{
					Tags: ClusterTag,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.SecurityGroupSelectorTerms()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.SecurityGroupSelectorTerms() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_UserData(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected *string
	}{
		{
			name: "Valid base64-encoded user data",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					UserData: lo.ToPtr("SGVsbG8sIHdvcmxkIQ=="),
				},
			},
			expected: lo.ToPtr("Hello, world!"),
		},
		{
			name: "Nil user data",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					UserData: nil,
				},
			},
			expected: nil,
		},
		{
			name: "Nil CustomLT",
			n: NodeGroup{
				CustomLT: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.UserData()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.UserData() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_BlockDeviceMappings(t *testing.T) {

	getDefaultBlockDeviceMappings := func(amiFamily *string, override bool) []*awskarpenter.BlockDeviceMapping {

		amiFamilyProvider := awskarpenterprovider.GetAMIFamily(amiFamily, &awskarpenterprovider.Options{})
		mappings := amiFamilyProvider.DefaultBlockDeviceMappings()
		if override {
			mappings[0].EBS.VolumeSize = k8sapiresource.NewQuantity(100*GiB, k8sapiresource.BinarySI)
		}
		return mappings
	}

	tests := []struct {
		name     string
		n        NodeGroup
		expected []*awskarpenter.BlockDeviceMapping
	}{
		{
			name: "Block device mappings in CustomLT",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					BlockDeviceMappings: []ec2types.LaunchTemplateBlockDeviceMapping{
						{
							DeviceName: lo.ToPtr("/dev/xvda"),
							Ebs: &ec2types.LaunchTemplateEbsBlockDevice{
								VolumeSize:          lo.ToPtr(int32(100)),
								VolumeType:          ec2types.VolumeTypeGp2,
								DeleteOnTermination: lo.ToPtr(true),
								Encrypted:           lo.ToPtr(false),
								Iops:                lo.ToPtr(int32(3000)),
								KmsKeyId:            lo.ToPtr("my-kms-key-id"),
								SnapshotId:          lo.ToPtr("my-snapshot-id"),
								Throughput:          lo.ToPtr(int32(500)),
							},
						},
					},
				},
			},
			expected: []*awskarpenter.BlockDeviceMapping{
				{
					DeviceName: lo.ToPtr("/dev/xvda"),
					EBS: &awskarpenter.BlockDevice{
						VolumeSize:          k8sapiresource.NewQuantity(100*GiB, k8sapiresource.BinarySI),
						VolumeType:          lo.ToPtr("gp2"),
						DeleteOnTermination: lo.ToPtr(true),
						Encrypted:           lo.ToPtr(false),
						IOPS:                lo.ToPtr(int64(3000)),
						KMSKeyID:            lo.ToPtr("my-kms-key-id"),
						SnapshotID:          lo.ToPtr("my-snapshot-id"),
						Throughput:          lo.ToPtr(int64(500)),
					},
				},
			},
		},
		{
			name: "Managed node group with Linux custom disk size",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType:  ekstypes.AMITypesAl2X8664,
					DiskSize: lo.ToPtr(int32(100)),
				},
			},
			expected: getDefaultBlockDeviceMappings(lo.ToPtr(awskarpenter.AMIFamilyAL2), true),
		},
		{
			name: "Managed node group with Windows custom disk size",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType:  ekstypes.AMITypesWindowsCore2019X8664,
					DiskSize: lo.ToPtr(int32(100)),
				},
			},
			expected: getDefaultBlockDeviceMappings(lo.ToPtr(awskarpenter.AMIFamilyWindows2019), true),
		},
		{
			name: "Managed node group with Linux default disk size",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType:  ekstypes.AMITypesAl2023X8664Standard,
					DiskSize: lo.ToPtr(ALAndBottleRocketDefaultDiskSize),
				},
			},
			expected: getDefaultBlockDeviceMappings(lo.ToPtr(awskarpenter.AMIFamilyAL2023), false),
		},
		{
			name: "Managed node group with Windows default disk size",
			n: NodeGroup{
				Nodegroup: &ekstypes.Nodegroup{
					AmiType:  ekstypes.AMITypesWindowsFull2022X8664,
					DiskSize: lo.ToPtr(WindowsDefaultDiskSize),
				},
			},
			expected: getDefaultBlockDeviceMappings(lo.ToPtr(awskarpenter.AMIFamilyWindows2022), false),
		},
		{
			name: "Nil CustomLT",
			n: NodeGroup{
				CustomLT: nil,
				Nodegroup: &ekstypes.Nodegroup{
					AmiType:  ekstypes.AMITypesBottlerocketX8664,
					DiskSize: lo.ToPtr(ALAndBottleRocketDefaultDiskSize),
				},
			},
			expected: getDefaultBlockDeviceMappings(lo.ToPtr(awskarpenter.AMIFamilyBottlerocket), false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.BlockDeviceMappings()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.BlockDeviceMappings() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNodeGroup_MetadataOptions(t *testing.T) {
	tests := []struct {
		name     string
		n        NodeGroup
		expected *awskarpenter.MetadataOptions
	}{
		{
			name: "Valid metadata options in CustomLT",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					MetadataOptions: &ec2types.LaunchTemplateInstanceMetadataOptions{
						HttpEndpoint:            ec2types.LaunchTemplateInstanceMetadataEndpointStateEnabled,
						HttpPutResponseHopLimit: lo.ToPtr(int32(2)),
						HttpTokens:              ec2types.LaunchTemplateHttpTokensStateRequired,
						HttpProtocolIpv6:        ec2types.LaunchTemplateInstanceMetadataProtocolIpv6Enabled,
					},
				},
			},
			expected: &awskarpenter.MetadataOptions{
				HTTPEndpoint:            lo.ToPtr("enabled"),
				HTTPPutResponseHopLimit: lo.ToPtr(int64(2)),
				HTTPTokens:              lo.ToPtr("required"),
				HTTPProtocolIPv6:        lo.ToPtr("enabled"),
			},
		},
		{
			name: "Nil metadata options in CustomLT",
			n: NodeGroup{
				CustomLT: &ec2types.ResponseLaunchTemplateData{
					MetadataOptions: nil,
				},
			},
			expected: nil,
		},
		{
			name: "Nil CustomLT",
			n: NodeGroup{
				CustomLT: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.MetadataOptions()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NodeGroup.MetadataOptions() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
