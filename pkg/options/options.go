package options

import (
	"github.com/spf13/cobra"

	"github.com/punkwalker/karpenter-generate/pkg/k8s"
)

type Options struct {
	ClusterName   string
	NodegroupName string
	Profile       string
	Region        string
	Debug         bool
}

func New(cmd *cobra.Command) *Options {
	opts := Options{}
	cmd.Flags().StringVar(&opts.Profile, "profile", "", "use the specific profile from your credential file")
	cmd.Flags().StringVar(&opts.Region, "region", "", "the region to use, overrides config/env settings")
	cmd.Flags().BoolVar(&opts.Debug, "debug", opts.Debug, "")
	cmd.Flags().MarkHidden("debug")
	cmd.Flags().StringVar(&opts.ClusterName, "cluster", "", "name of the EKS cluster")
	cmd.Flags().StringVar(&opts.NodegroupName, "nodegroup", "", "name of the EKS managed nodegroup")
	cmd.SetHelpFunc(usage)

	return &opts
}

func (o *Options) Parse() {
	if o.ClusterName == "" && o.Region == "" {
		o.ClusterName, o.Region = k8s.OptionsFromConfig()
	}
}

func usage(cmd *cobra.Command, args []string) {
	usageString := `
Description:
  A CLI tool to generate Karpenter Custom Resources such as
  Nodepools and EC2NodeClass from details of EKS Managed Nodegroup

Usage:
  karpenter-generate [flags]

Optional Flags:
  --cluster string     name of the EKS cluster 
                       (default: from kubeconfig current-context)
  --nodegroup string   name of the EKS managed nodegroup 
                       (default: all the nodegroups)
  --region string      the region to use, overrides config/env settings 
                       (default: from kubeconfig current-context or AWS config)
  --profile string     use the specific profile from your credential file 
                       (default: from kubeconfig current-context or AWS config)
  -h, --help           help for karpenter-generate
	`
	cmd.Println(usageString)
}
