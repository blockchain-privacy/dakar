package user

import (
	"backend/user"
	"fmt"
	"regexp"
	"time"
)

const (
	DTypeUser = "User"
	DTypeRole = "Role"
)

type Role struct {
	Uid   string   `json:"uid,omitempty"`
	Name  string   `json:"role_name"`
	DType []string `json:"dgraph.type,omitempty"`
}

func (r Role) String() string {
	return fmt.Sprintf("uid: %s, name: %s", r.Uid, r.Name)
}

func (r *Role) SetDType() {
	r.DType = []string{DTypeRole}
}

type User struct {
	Uid          string    `json:"uid,omitempty"`
	Email        string    `json:"user_email"`
	PasswordHash string    `json:"user_pwhash"`
	Roles        []Role    `json:"user_roles"`
	Created      time.Time `json:"user_created"`
	Modified     time.Time `json:"user_modified"`
	DType        []string  `json:"dgraph.type,omitempty"`
}

func (u User) String() string {
	return fmt.Sprintf("uid: %s, email %s, roles: %v, created: %s, modified: %s",
		u.Uid, u.Email, u.Roles, u.Created, u.Modified)
}

func (u *User) SetDType() {
	u.DType = []string{DTypeUser}
}

type FrontendUserCreate struct {
	Email string   `json:"user_email"`
	Roles []string `json:"user_roles"`
}

func (f FrontendUserCreate) String() string {
	return fmt.Sprintf("email %s, roles: %v", f.Email, f.Roles)
}

func (f FrontendUserCreate) ToUser() User {
	var roles []Role

	for _, r := range f.Roles {
		roles = append(roles, Role{
			Name: r,
		})
	}

	if len(roles) == 0 {
		roles = nil
	}

	return User{
		Email: f.Email,
		Roles: roles,
	}
}

// isValidEmail is a regex filter which checks if the input conforms to an email string
var isValidEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]" +
	"{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$").MatchString

// IsValid does a sanity check for the given FrontendUserCreate
func (f FrontendUserCreate) IsValid() bool {
	// check if values are set
	if len(f.Email) == 0 || len(f.Roles) == 0 || !isValidEmail(f.Email) {
		return false
	}

	// check if all roles have valid values
	for _, ur := range f.Roles {
		if _, err := user.GetRoleByName(ur); err != nil {
			return false
		}
	}

	return true
}

type FrontendUserLogin struct {
	Email    string `json:"user_email"`
	Password string `json:"user_pw"`
}

func (f FrontendUserLogin) String() string {
	return fmt.Sprintf("email %s, pw: %s", f.Email, f.Password)
}

// IsValid does a sanity check for the given FrontendUserLogin
func (f FrontendUserLogin) IsValid() bool {
	return len(f.Email) > 0 || len(f.Password) > 0
}
