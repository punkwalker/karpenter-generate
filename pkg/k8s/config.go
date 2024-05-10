package k8s

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func Client(context string) (*kubernetes.Clientset, error) {
	var clientConfigOverrides *clientcmd.ConfigOverrides
	if context != "" {
		clientConfigOverrides = &clientcmd.ConfigOverrides{
			CurrentContext: context,
		}
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		clientConfigOverrides,
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, genericclioptions.ErrEmptyConfig
	}

	return kubernetes.NewForConfig(restConfig)
}

func OptionsFromConfig() (string, string, string) {
	var clusterName string
	var region string
	var profile string
	raw, err := Kubeconfig()
	if err != nil {
		return clusterName, region, profile
	}

	if err := clientcmdapi.MinifyConfig(raw); err != nil {
		return clusterName, region, profile
	}

	for _, user := range raw.AuthInfos {
		for idx, val := range user.Exec.Args {
			switch val {
			case "-i", "--cluster-name":
				clusterName = user.Exec.Args[idx+1]
			case "--region":
				region = user.Exec.Args[idx+1]
			}
		}

		for _, env := range user.Exec.Env {
			if env.Name == "AWS_REGION" {
				region = env.Value
			}
			if env.Name == "AWS_PROFILE" {
				profile = env.Value
			}
		}
	}
	return clusterName, region, profile
}

func ClusterURLForCurrentContext() string {
	raw, err := Kubeconfig()
	if err != nil {
		return ""
	}

	if err := clientcmdapi.MinifyConfig(raw); err != nil {
		return ""
	}

	return raw.Clusters[raw.Contexts[raw.CurrentContext].Cluster].Server
}

func DynamicClient(context string) (dynamic.Interface, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		},
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(restConfig)
}

func KubeContextForCluster(cluster *types.Cluster) (string, error) {
	raw, err := Kubeconfig()
	if err != nil {
		return "", err
	}

	found := ""

	for name, context := range raw.Contexts {
		if _, ok := raw.Clusters[context.Cluster]; ok {
			if raw.Clusters[context.Cluster].Server == aws.ToString(cluster.Endpoint) {
				found = name
				break
			}
		}
	}

	return found, nil
}

func Kubeconfig() (*clientcmdapi.Config, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	raw, err := config.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return &raw, nil
}
