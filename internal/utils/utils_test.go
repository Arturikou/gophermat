package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "Valid number",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "Valid number",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "Invalid number",
			number: "12345678904",
			want:   false,
		},
		{
			name:   "Invalid characters",
			number: "1234567890A",
			want:   false,
		},
		{
			name:   "Special characters",
			number: "12345-67890",
			want:   false,
		},
		{
			name:   "Empty string",
			number: "",
			want:   true,
		},
		{
			name:   "Single digit valid",
			number: "0",
			want:   true,
		},
		{
			name:   "Single digit invalid",
			number: "5",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidNumber(tt.number)
			assert.Equal(t, tt.want, got, "IsValidNumber(%s) should be %v", tt.number, tt.want)
		})
	}
}
