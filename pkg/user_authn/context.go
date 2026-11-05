package userauthn

import "context"

type userIDKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(userIDKey{}).(string)
	return uid, ok
}

func contextWithUserID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userIDKey{}, uid)
}
