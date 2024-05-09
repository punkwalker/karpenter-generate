package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/punkwalker/karpenter-generate/pkg/aws"
	"github.com/punkwalker/karpenter-generate/pkg/karpenter"
	"github.com/punkwalker/karpenter-generate/pkg/options"
)

func run(opts *options.Options) error {
	aws.Init(opts)
	opts.Parse()
	var nodeGroups []types.Nodegroup

	eksClient := aws.NewEKSClient()
	if opts.NodegroupName != "" {
		nodeGroup, err := eksClient.DescribeNodegroup(opts.ClusterName, opts.NodegroupName)
		if err != nil {
			return aws.FormatErrorAsMessageOnly(err)
		}
		nodeGroups = append(nodeGroups, *nodeGroup)
		nodePools, nodeClasses, err := karpenter.Generate(&nodeGroups)
		if err != nil {
			return err
		}
		return karpenter.Print(nodePools, nodeClasses)
	}

	nodeGroups, err := eksClient.GetAllNodeGroups(opts)
	if err != nil {
		return aws.FormatErrorAsMessageOnly(err)
	}

	if len(nodeGroups) == 0 {
		return fmt.Errorf("no nodegroups found")
	}

	nodePools, nodeClasses, err := karpenter.Generate(&nodeGroups)
	if err != nil {
		return err
	}
	return karpenter.Print(nodePools, nodeClasses)
}
