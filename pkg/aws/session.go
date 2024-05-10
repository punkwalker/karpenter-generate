package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/punkwalker/karpenter-generate/pkg/options"
)

var awsConfig *aws.Config
var profile string
var region string
var debug bool

func Init(opts *options.Options) {
	profile = opts.Profile
	region = opts.Region
	debug = opts.Debug
}

func GetConfig() aws.Config {
	if awsConfig != nil {
		return *awsConfig
	}

	cfgOptions := []func(*config.LoadOptions) error{
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	}

	if debug {
		logMode := aws.LogRetries | aws.LogRequestWithBody

		cfgOptions = append(cfgOptions,
			config.WithClientLogMode(logMode),
		)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), cfgOptions...)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create AWS config: %w", err))
	}
	region = cfg.Region
	awsConfig = &cfg

	return cfg
}
