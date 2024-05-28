package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

type EKSClient struct {
	*eks.Client
}

func NewEKSClient() *EKSClient {
	return &EKSClient{eks.NewFromConfig(GetConfig())}
}

func (c *EKSClient) ListNodegroups(clusterName string) ([]string, error) {
	nodegroupNames := []string{}
	pageNum := 0

	paginator := eks.NewListNodegroupsPaginator(c.Client, &eks.ListNodegroupsInput{
		ClusterName: aws.String(clusterName),
	})

	for paginator.HasMorePages() && pageNum < maxPages {
		out, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		nodegroupNames = append(nodegroupNames, out.Nodegroups...)
		pageNum++
	}

	return nodegroupNames, nil
}

func (c *EKSClient) DescribeNodegroup(clusterName, nodegroupName string) (*types.Nodegroup, error) {
	result, err := c.Client.DescribeNodegroup(context.Background(), &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodegroupName),
	})

	if err != nil {
		return nil, err
	}

	return result.Nodegroup, nil
}
