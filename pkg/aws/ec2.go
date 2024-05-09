package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2Client struct {
	*ec2.Client
}

func NewEC2Client() *EC2Client {
	return &EC2Client{ec2.NewFromConfig(GetConfig())}
}

// Describes the specified Launch Template versions or all of your Launch Template versions.
func (c *EC2Client) DescribeLaunchTemplateVersions(id, version string) ([]types.LaunchTemplateVersion, error) {
	pageNum := 0
	versions := []types.LaunchTemplateVersion{}
	input := ec2.DescribeLaunchTemplateVersionsInput{}

	if id != "" {
		input.LaunchTemplateId = &id
	}

	if version != "" {
		input.Versions = []string{version}
	}

	paginator := ec2.NewDescribeLaunchTemplateVersionsPaginator(c.Client, &input)

	for paginator.HasMorePages() && pageNum < maxPages {
		out, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		versions = append(versions, out.LaunchTemplateVersions...)
		pageNum++
	}

	return versions, nil
}

func (c *EC2Client) DescribeVolumes(id string) ([]types.Volume, error) {
	filters := []types.Filter{}
	volumes := []types.Volume{}
	pageNum := 0

	if id != "" {
		filters = append(filters, types.Filter{
			Name:   aws.String("volume-id"),
			Values: []string{id},
		})
	}

	input := ec2.DescribeVolumesInput{}
	if len(filters) > 0 {
		input.Filters = filters
	}

	paginator := ec2.NewDescribeVolumesPaginator(c.Client, &input)

	for paginator.HasMorePages() && pageNum < maxPages {
		out, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, out.Volumes...)
		pageNum++
	}

	return volumes, nil
}
