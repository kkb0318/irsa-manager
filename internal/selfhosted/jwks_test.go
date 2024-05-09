package selfhosted

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const rsaKeyID = "JHJehTTTZlsspKHT-GaJxK7Kd1NQgZJu3fyK6K_QDYU"

func TestJWK(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		expected  string
		expectErr bool
	}{
		{
			name:     "rsa",
			filename: "testdata/rsa.pub",
			expected: rsaKeyID,
		},
		{
			name:      "no rsa",
			filename:  "testdata/ecdsa.pub",
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.filename)
			assert.NoError(t, err)
			actual, err := NewJWK(content)
			if tt.expectErr {
				assert.Error(t, err, "")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual.Keys[0].KeyID)
			}
		})
	}
}
