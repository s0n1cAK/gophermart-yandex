package luhn

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid number",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "invalid number",
			number: "79927398714",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "non-digit character",
			number: "7992a398713",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Valid(tt.number)
			require.Equal(t, tt.want, got)
		})
	}
}
