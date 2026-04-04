package ctxkeys

import (
	"context"
	"errors"
)

var ErrContextKeyNotFound = errors.New("key not found in context")
var ErrContextValueWrongType = errors.New("context value has unexpected type")

type contextKey string

const keyUserID contextKey = "userID"

func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, keyUserID, userID)
}

func UserIDFromContext(ctx context.Context) (int, error) {
	val := ctx.Value(keyUserID)
	if val == nil {
		return 0, ErrContextKeyNotFound
	}

	id, ok := val.(int)
	if !ok {
		return 0, ErrContextValueWrongType
	}

	return id, nil
}
