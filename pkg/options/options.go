package options

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Options struct {
	ClusterName            string
	NodegroupName          string
	KarpenterNodegroupName string
	Profile                string
	Region                 string
	Output                 string
	Debug                  bool
}

func New(cmd *cobra.Command) *Options {
	opts := Options{}
	cmd.Flags().StringVar(&opts.Profile, "profile", "", "use the specific profile from your credential file")
	cmd.Flags().StringVar(&opts.Region, "region", "", "the region to use, overrides config/env settings")
	cmd.Flags().BoolVar(&opts.Debug, "debug", opts.Debug, "")
	cmd.Flags().StringVar(&opts.ClusterName, "cluster", "", "name of the EKS cluster")
	cmd.Flags().StringVar(&opts.NodegroupName, "nodegroup", "", "name of the EKS managed nodegroup")
	cmd.Flags().StringVar(&opts.KarpenterNodegroupName, "karpenter-nodegroup", "", "name of the EKS managed nodegroup running Karpenter deployment")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "yaml", "name of the EKS managed nodegroup running Karpenter deployment")
	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("karpenter-nodegroup")
	_ = cmd.Flags().MarkHidden("debug")
	cmd.SetHelpFunc(usage)

	return &opts
}

func (o *Options) Parse() error {
	if o.ClusterName == "" {
		return fmt.Errorf(`specify value for "--cluster" flag (e.g.: karpenter-generate --cluster <Cluster Name> --karpenter-nodegroup <Karpenter Nodegroup Name>)`)
	}
	if o.KarpenterNodegroupName == "" {
		return fmt.Errorf(`specify value for "--karpenter-nodegroup" flag (e.g.: karpenter-generate --cluster <Cluster Name> --karpenter-nodegroup <Karpenter Nodegroup Name>)`)
	}
	return nil
}

func usage(cmd *cobra.Command, _ []string) {
	usageString := `
Description:
  A CLI tool to generate Karpenter Custom Resources such as
  Nodepools and EC2NodeClass from details of EKS Managed Nodegroup

Usage:
  karpenter-generate --cluster <Cluster Name> --karpenter-nodegroup <Karpenter Nodegroup Name>

Available Commands:
  version     Print the version and build information for karpenter-generate

Flags:
  --cluster string               name of the EKS cluster 
  --karpenter-nodegroup string   name of the EKS managed nodegroup running Karpenter deployment or fargate
									 
Optiona Flags:
  --nodegroup string   name of the EKS managed nodegroup 
                       (default: all the nodegroups expectthe one running Karpenter)
  --region string      region of EKS cluster, overrides AWS CLI configuration/ENV values 
                       (default: AWS CLI configuration)
  --profile string     use the specific profile from your credential file 
                       (default: AWS CLI configuration)
  --output string      output format (yaml or json)
					   (default: yaml)
  -h, --help           help for karpenter-generate
	`
	cmd.Println(usageString)
}
