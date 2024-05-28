package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		opts    *Options
		wantErr bool
	}{
		{
			name: "Valid options",
			opts: &Options{
				ClusterName:            "my-cluster",
				KarpenterNodegroupName: "my-karpenter-nodegroup",
			},
			wantErr: false,
		},
		{
			name: "Missing cluster name",
			opts: &Options{
				ClusterName:            "",
				KarpenterNodegroupName: "my-karpenter-nodegroup",
			},
			wantErr: true,
		},
		{
			name: "Missing karpenter nodegroup name",
			opts: &Options{
				ClusterName:            "my-cluster",
				KarpenterNodegroupName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Parse()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
