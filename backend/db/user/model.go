package user

import "slices"

// DType is the dgraph database type for the User type
const DType = "User"

// User is the database representation of a user
type User struct {
	UID   string   `json:"uid,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (u *User) SetDType() {
	u.DType = []string{DType}
}

var KratosAllowedIdentityStates = []string{
	"active",
	"inactive",
}

func IsStateValid(state string) bool {
	return slices.Contains(KratosAllowedIdentityStates, state)
}
