package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/punkwalker/karpenter-generate/pkg/aws"
	"github.com/punkwalker/karpenter-generate/pkg/karpenteraws"
	"github.com/punkwalker/karpenter-generate/pkg/printers"
)

func run() error {
	aws.Init(opts)
	if err := opts.Parse(); err != nil {
		return err
	}

	printer, err := printers.NewPrinter(printers.Output(opts.Output))
	if err != nil {
		return err
	}

	nodeGroups, err := getNodegroups()
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

func getNodegroups() ([]types.Nodegroup, error) {
	var nodegroups []types.Nodegroup
	var ngList []string
	var err error

	eksClient := aws.NewEKSClient()

	if opts.NodegroupName != "" {
		ngList = []string{opts.NodegroupName}
	} else {
		ngList, err = eksClient.ListNodegroups(opts.ClusterName)
		if err != nil {
			return nil, err
		}
	}

	for _, ng := range ngList {
		if opts.KarpenterNodegroupName != ng {
			nodegroup, err := eksClient.DescribeNodegroup(opts.ClusterName, ng)
			if err != nil {
				return nil, err
			}

			if nodegroup.Status != "ACTIVE" {
				return nil, fmt.Errorf(`nodegroup "%s" is not active, make sure all the nodegroups are in "ACTIVE" state`, ng)
			}
			nodegroups = append(nodegroups, *nodegroup)
		}
	}
	return nodegroups, nil
}
