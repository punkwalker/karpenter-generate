package printers

import (
	"fmt"
	"os"

	awskarpenter "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	"k8s.io/cli-runtime/pkg/printers"
	sigkarpenter "sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

type Output string

type OutputPrinter struct {
	printers.ResourcePrinter
}

const (
	YAML Output = "yaml"
	JSON Output = "json"
)

func NewPrinter(format Output) (printers.ResourcePrinter, error) {
	switch format {
	case YAML:
		return &printers.YAMLPrinter{}, nil
	case JSON:
		return &printers.JSONPrinter{}, nil
	default:
		return nil, fmt.Errorf(`invalid output type, valid values are "yaml" or "json"`)
	}
}

func Print(p printers.ResourcePrinter, nodePools []sigkarpenter.NodePool, nodeClasses []awskarpenter.EC2NodeClass) error {
	if len(nodePools) != len(nodeClasses) {
		return fmt.Errorf("no. of nodepools is not equal to no. of nodeclass")
	}
	for idx := range len(nodePools) {
		err := p.PrintObj(&nodePools[idx], os.Stdout)
		if err != nil {
			return err
		}
		err = p.PrintObj(&nodeClasses[idx], os.Stdout)
		if err != nil {
			return err
		}
	}
	return nil
}
