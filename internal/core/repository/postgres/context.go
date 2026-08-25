package core_postgres

import "context"

type txContextKey struct{}

func ContextWithDB(
	ctx context.Context,
	db DB,
) context.Context {
	return context.WithValue(ctx, txContextKey{}, db)
}

func DBFromContext(
	ctx context.Context,
) DB {
	db, ok := ctx.Value(txContextKey{}).(DB)
	if !ok {
		return nil
	}
	return db
}
