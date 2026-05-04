package auth

import "context"

type contextKey string

const userClaimsKey contextKey = "user_claims"

func ContextWithUserClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

func UserClaimsFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	return claims, ok
}
