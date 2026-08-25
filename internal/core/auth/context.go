package core_auth

import "context"

type contextKey string

const AuthInfoKey contextKey = "auth_info"

func WithAuthInfo(
	ctx context.Context,
	info AuthInfo,
) context.Context {
	return context.WithValue(
		ctx,
		AuthInfoKey,
		info,
	)
}

func AuthInfoFromContext(
	ctx context.Context,
) (AuthInfo, bool) {
	info, ok := ctx.Value(AuthInfoKey).(AuthInfo)
	return info, ok
}
