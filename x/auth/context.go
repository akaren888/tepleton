package auth

import (
	"github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/x/store"
)

/*

	Usage:

	import "accounts"

	var acc accounts.Account

	accounts.SetAccount(ctx, acc)
	acc2 := accounts.GetAccount(ctx)

*/

type contextKey int // local to the auth module

const (
	// A context key of the Account variety
	contextKeyAccount contextKey = iota
)

func SetAccount(ctx types.Context, account store.Account) types.Context {
	return ctx.WithValueUnsafe(contextKeyAccount, account)
}

func GetAccount(ctx types.Context) store.Account {
	return ctx.Value(contextKeyAccount).(store.Account)
}
