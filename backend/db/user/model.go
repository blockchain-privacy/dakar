package user

import (
	"fmt"
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
	return fmt.Sprintf("uid: %s, name %s", r.Uid, r.Name)
}

func (r *Role) SetDType() {
	r.DType = []string{DTypeRole}
}

type User struct {
	Uid      string    `json:"uid,omitempty"`
	Email    string    `json:"user_email"`
	Roles    []Role    `json:"user_roles"`
	Created  time.Time `json:"user_created"`
	Modified time.Time `json:"user_modified"`
	DType    []string  `json:"dgraph.type,omitempty"`
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
