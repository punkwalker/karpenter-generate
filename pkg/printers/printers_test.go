package printers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/printers"
)

func TestNewPrinter(t *testing.T) {
	tests := []struct {
		name     string
		format   Output
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "YAML Printer",
			format:   YAML,
			expected: &printers.YAMLPrinter{},
			wantErr:  false,
		},
		{
			name:     "JSON Printer",
			format:   JSON,
			expected: &printers.JSONPrinter{},
			wantErr:  false,
		},
		{
			name:     "Invalid Format",
			format:   "invalid",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer, err := NewPrinter(tt.format)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.IsType(t, tt.expected, printer)
			}
		})
	}
}
