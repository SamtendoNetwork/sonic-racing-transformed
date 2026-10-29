package globals

import (
	"fmt"
	"strconv"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
)



func AccountDetailsByPID(pid types.PID) (*nex.Account, *nex.Error) {
	if AuthenticationServerAccount != nil && pid.Equals(AuthenticationServerAccount.PID) {
		return AuthenticationServerAccount, nil
	}

	if SecureServerAccount != nil && pid.Equals(SecureServerAccount.PID) {
		return SecureServerAccount, nil
	}

	password, errorCode := PasswordFromPID(pid)
	if errorCode != 0 {
		return nil, nex.NewError(errorCode, "Accounts password lookup failed")
	}


	return nex.NewAccount(pid, strconv.FormatUint(uint64(pid), 10), password, false), nil
}




func AccountDetailsByUsername(username string) (*nex.Account, *nex.Error) {
	if AuthenticationServerAccount != nil && username == AuthenticationServerAccount.Username {
		return AuthenticationServerAccount, nil
	}

	if SecureServerAccount != nil && username == SecureServerAccount.Username {
		return SecureServerAccount, nil
	}

	pid, err := strconv.ParseUint(username, 10, 64)
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidUsername, fmt.Sprintf("Username %q is not a valid NEX PID", username))
	}

	password, errorCode := PasswordFromPID(types.NewPID(pid))
	if errorCode != 0 {
		return nil, nex.NewError(errorCode, "Accounts password lookup failed")
	}

	return nex.NewAccount(types.NewPID(pid), username, password, false), nil
}
