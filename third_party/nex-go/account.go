package nex

import "github.com/PretendoNetwork/nex-go/v2/types"








type Account struct {
	PID               types.PID
	Username          string
	Password          string
	RequiresTokenAuth bool
}




func NewAccount(pid types.PID, username, password string, requiresTokenAuth bool) *Account {
	return &Account{
		PID:               pid,
		Username:          username,
		Password:          password,
		RequiresTokenAuth: requiresTokenAuth,
	}
}
