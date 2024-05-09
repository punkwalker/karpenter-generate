package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/punkwalker/karpenter-generate/pkg/options"
)

var opts *options.Options

var rootCmd = &cobra.Command{
	Use:   "karpenter-generate",
	Short: "Tool to generate Karpenter CRDs from EKS Managed Nodegroups",
	Long: `This is a CLI tool which can be used to generate Karpenter Custom Resources such as 
Nodepools and EC2NodeClass from details of EKS Managed Nodegroup. Which will allow seamless migration to Karpenter.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(opts)
	},
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
