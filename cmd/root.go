package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/punkwalker/karpenter-generate/pkg/aws"
	"github.com/punkwalker/karpenter-generate/pkg/karpenteraws"
	"github.com/punkwalker/karpenter-generate/pkg/options"
	"github.com/punkwalker/karpenter-generate/pkg/printers"
)

var opts *options.Options

var rootCmd = &cobra.Command{
	Use:   "karpenter-generate",
	Short: "Tool to generate Karpenter CRDs from EKS Managed Nodegroups",
	Long: `This is a CLI tool which can be used to generate Karpenter Custom Resources such as 
Nodepools and EC2NodeClass from details of EKS Managed Nodegroup. Which will allow seamless migration to Karpenter.`,
	SilenceUsage: true,
	RunE:         run,
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	opts = options.New(rootCmd)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

func run(_ *cobra.Command, _ []string) error {
	aws.Init(opts)
	if err := opts.Parse(); err != nil {
		return err
	}

	printer, err := printers.NewPrinter(printers.Output(opts.Output))
	if err != nil {
		return err
	}

	nodePools, nodeClasses, err := karpenteraws.Generate(opts)
	if err != nil {
		return err
	}
	return printers.Print(printer, nodePools, nodeClasses)
}
