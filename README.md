# `karpenter-generate`
This is a simple CLI tool to generate AWS Karpenter Custom Kubernetes Resources (Nodepool & EC2NodeClass) from AWS EKS Managed Nodegroup information. The generated resources can be stored in as a yaml manifest file or can be directly applied to the cluster.

## Example Usage
### For All Managed Nodegroups
The tool can be used without any flag and it will scan all the Managed Nodegroups in the cluster. The cluster to scan is decided based on current-context in kubeconfig
```bash
karpenter-generate
```
### For specific Managed Nodegroup
To generate Karpenter Custom Resources for a specific Managed Nodegroup.
```bash
karpenter-generate --nodegroup <Managed_Nodegroup_Name>
```

### For cluster other than current-context
To specifcy other cluster than current-context
```bash
karpenter-generate  --cluster <Cluster_Name>  --region <Region>
```

## Prerequisites
- Configured [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- Valid kubeconfig store at `$HOME/.kube/config`

## Installation
### MacOS & Linux
Use [Homebrew](https://brew.sh/) and run following command.
```bash
brew tap punkwalker/tap
brew install karpenter-generate
```

### Manually (Linux/Windows)
Downloaded archive file from release artifacts. Download the archive file from relase page according to the Architecture of your machine

| OS | Arch | Download|
| ------ | ------ | ------ |
| Linux   | AMD64/x86_64 | [Link](https://github.com/punkwalker/karpenter-generate/releases/download/v0.0.1/karpenter-generate_Linux_x86_64.tar.gz)|
|    | ARM64| [Link](https://github.com/punkwalker/karpenter-generate/releases/download/v0.0.3/karpenter-generate_Linux_arm64.tar.gz)|
| Windows   | AMD64/x86_64 | [Link](https://github.com/punkwalker/karpenter-generate/releases/download/v0.0.1/karpenter-generate_Windows_x86_64.tar.gz)|
|    | ARM64| [Link](https://github.com/punkwalker/karpenter-generate/releases/download/v0.0.3/karpenter-generate_Windows_arm64.tar.gz)|

After downloading the archive, extract it and copy the binary/executable to `/usr/local/bin` for Linux. For Windows, run the `karpenter-generate.exe` from extracted folder.

## Help
```bash
karpeter-generate  --help

Description:
  A CLI tool to generate Karpenter Custom Resources such as
  Nodepools and EC2NodeClass from details of EKS Managed Nodegroup

Usage:
  karpenter-generate [command]
  karpenter-generate [flags]

Available Commands:
  version     Print the version and build information for karpenter-generate

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
```

## Contributing
Contributions are welcome! If you encounter any issues or have suggestions for improvements, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them with descriptive commit messages.
4. Push your changes to your forked repository.
5. Submit a pull request to the main repository.

Please ensure that your code adheres to the project's coding standards and includes appropriate tests.

## License
This tool is licensed under the License Name License. See the `LICENSE.md` file for more information.