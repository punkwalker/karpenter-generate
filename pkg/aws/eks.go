package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/punkwalker/karpenter-generate/pkg/options"
)

type EKSClient struct {
	*eks.Client
}

func NewEKSClient() *EKSClient {
	return &EKSClient{eks.NewFromConfig(GetConfig())}
}

func (c *EKSClient) GetAllNodeGroups(opts *options.Options) ([]types.Nodegroup, error) {
	var nodegroups []types.Nodegroup
	ngList, err := c.ListNodegroups(opts.ClusterName)

	if err != nil {
		return nil, err
	}

	for _, ng := range ngList {
		nodegroup, err := c.DescribeNodegroup(opts.ClusterName, ng)
		if err != nil {
			return nil, err
		}

		nodegroups = append(nodegroups, *nodegroup)
	}

	return nodegroups, nil
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
