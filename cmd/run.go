package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/punkwalker/karpenter-generate/pkg/aws"
	"github.com/punkwalker/karpenter-generate/pkg/karpenteraws"
	"github.com/punkwalker/karpenter-generate/pkg/options"
	"github.com/punkwalker/karpenter-generate/pkg/printers"
)

func run(opts *options.Options) error {
	aws.Init(opts)
	if err := opts.Parse(); err != nil {
		return err
	}
	var nodeGroups []types.Nodegroup

	eksClient := aws.NewEKSClient()
	printer, err := printers.NewPrinter("yaml")
	if err != nil {
		return err
	}

	if opts.NodegroupName != "" {
		nodeGroup, err := eksClient.DescribeNodegroup(opts.ClusterName, opts.NodegroupName)
		if err != nil {
			return aws.FormatErrorAsMessageOnly(err)
		}
		nodeGroups = append(nodeGroups, *nodeGroup)

		nodePools, nodeClasses, err := karpenteraws.Generate(&nodeGroups)
		if err != nil {
			return err
		}

		return printers.Print(printer, nodePools, nodeClasses)
	}

	nodeGroups, err = eksClient.GetAllNodeGroups(opts)
	if err != nil {
		return aws.FormatErrorAsMessageOnly(err)
	}

	if len(nodeGroups) == 0 {
		return fmt.Errorf("no nodegroups found")
	}

	nodePools, nodeClasses, err := karpenteraws.Generate(&nodeGroups)
	if err != nil {
		return err
	}
	return printers.Print(printer, nodePools, nodeClasses)
}
