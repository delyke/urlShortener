package appctx

import "context"

type userIDKeyType struct{}

var userIDKey = userIDKeyType{}

func SetUserID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func UserID(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
