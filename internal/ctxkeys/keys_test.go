package ctxkeys

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextUserID(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedID    int
		expectedError error
	}{
		{
			name: "get existing userID",
			setupContext: func() context.Context {
				return SetUserID(context.Background(), 42)
			},
			expectedID:    42,
			expectedError: nil,
		},
		{
			name: "key not found",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedID:    0,
			expectedError: ErrContextKeyNotFound,
		},
		{
			name: "wrong value type",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), keyUserID, "not-an-int")
			},
			expectedID:    0,
			expectedError: ErrContextValueWrongType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()

			id, err := UserIDFromContext(ctx)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Equal(t, 0, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}
