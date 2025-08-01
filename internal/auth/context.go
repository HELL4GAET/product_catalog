package auth

import "context"

type ctxKey string

const (
	RoleCtxKey   ctxKey = "role"
	UserIDCtxKey ctxKey = "userID"
)

func WithUserContext(ctx context.Context, userID int, role Role) context.Context {
	ctx = context.WithValue(ctx, UserIDCtxKey, userID)
	ctx = context.WithValue(ctx, RoleCtxKey, role)
	return ctx
}

func RoleFromContext(ctx context.Context) (Role, bool) {
	role, ok := ctx.Value(RoleCtxKey).(Role)
	return role, ok
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(UserIDCtxKey).(int)
	return id, ok
}
