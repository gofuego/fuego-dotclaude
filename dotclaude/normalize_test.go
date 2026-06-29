package dotclaude

import (
	"reflect"
	"testing"

	"github.com/gofuego/fuego/core"
)

func TestNormalizeTools(t *testing.T) {
	tests := []struct {
		name string
		env  core.Envelope
		want []string
	}{
		{"comma string", core.Envelope{"tools": "Read, Grep, Bash"}, []string{"Read", "Grep", "Bash"}},
		{"yaml list", core.Envelope{"tools": []any{"Read", "Write"}}, []string{"Read", "Write"}},
		{"allowed-tools fallback", core.Envelope{"allowed-tools": "Bash, Edit"}, []string{"Bash", "Edit"}},
		{"tools wins over allowed-tools", core.Envelope{"tools": "Read", "allowed-tools": "Bash"}, []string{"Read"}},
		{"missing", core.Envelope{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeTools(tt.env)
			got, _ := tt.env["tools"].([]string)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeTools -> %v, want %v", got, tt.want)
			}
		})
	}
}
