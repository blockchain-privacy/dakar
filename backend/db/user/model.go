package user

import (
	"backend/db/analytics/heuristics/transaction"
	"backend/user"
	"fmt"
	"regexp"
	"time"
)

const (
	// DTypeUser is the dgraph database type for the User type
	DTypeUser = "User"
	// DTypeRole is the dgraph database type for the Role type
	DTypeRole = "Role"
)

// Role is the database representation of a role
type Role struct {
	UID   string   `json:"uid,omitempty"`
	Name  string   `json:"role_name"`
	DType []string `json:"dgraph.type,omitempty"`
}

func (r Role) String() string {
	return fmt.Sprintf("uid: %s, name: %s", r.UID, r.Name)
}

// SetDType sets the DType for dgraph type recognition
func (r *Role) SetDType() {
	r.DType = []string{DTypeRole}
}

type User struct {
	UID          string                  `json:"uid,omitempty"`
	Email        string                  `json:"user_email,omitempty"`
	PasswordHash string                  `json:"user_pwhash,omitempty"`
	Roles        []Role                  `json:"user_roles,omitempty"`
	Created      *time.Time              `json:"user_created,omitempty"`
	Modified     *time.Time              `json:"user_modified,omitempty"`
	Heuristics   []transaction.Heuristic `json:"user_heuristics,omitempty"`
	DType        []string                `json:"dgraph.type,omitempty"`
}

func (u User) String() string {
	return fmt.Sprintf("uid: %s, email %s, roles: %v, created: %s, modified: %s, heuristic count: %d",
		u.UID, u.Email, u.Roles, u.Created, u.Modified, len(u.Heuristics))
}

// SetDType sets the DType for dgraph type recognition
func (u *User) SetDType() {
	u.DType = []string{DTypeUser}
}

func (u User) ToFrontendUserState() FrontendUserClientState {
	return FrontendUserClientState{
		UID:   u.UID,
		Email: u.Email,
		Roles: u.Roles,
	}
}

func (u User) ToFrontendUserBackendState() FrontendUserBackendState {
	return FrontendUserBackendState{
		UID:      u.UID,
		Email:    u.Email,
		Roles:    u.Roles,
		Modified: u.Modified,
		Created:  u.Created,
	}
}

// ModifyUserRequest represents the client side state of the user
type ModifyUserRequest struct {
	UID             string `json:"uid,omitempty"`
	Email           string `json:"email,omitempty"`
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password,omitempty"`
	Roles           []Role `json:"roles,omitempty"`
}

func (m ModifyUserRequest) ToUser(pwHash string) User {
	return User{
		UID:          m.UID,
		Email:        m.Email,
		Roles:        m.Roles,
		PasswordHash: pwHash,
	}
}

// FrontendUserClientState represents the client side state of the user
type FrontendUserClientState struct {
	UID   string `json:"uid,omitempty"`
	Email string `json:"email,omitempty"`
	Roles []Role `json:"roles,omitempty"`
}

func (f FrontendUserClientState) ToUser() User {
	return User{
		UID:   f.UID,
		Email: f.Email,
		Roles: f.Roles,
	}
}

// IsValid does a sanity check for the given FrontendUserClientState
func (f FrontendUserClientState) IsValid() bool {
	// check if values are set
	if len(f.Email) == 0 || len(f.Roles) == 0 || !IsValidEmail(f.Email) {
		return false
	}

	// check if all roles have valid values
	for _, ur := range f.Roles {
		if _, err := user.GetRoleByName(ur.Name); err != nil {
			return false
		}
	}

	return true
}

type FrontendUserRoles struct {
	Email string   `json:"user_email"`
	Roles []string `json:"user_roles"`
}

func (f FrontendUserRoles) String() string {
	return fmt.Sprintf("email %s, roles: %v", f.Email, f.Roles)
}

func (f FrontendUserRoles) ToUser() User {
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

// IsValidEmail is a regex filter which checks if the input conforms to an email string
var IsValidEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]" +
	"{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$").MatchString

// IsValid does a sanity check for the given FrontendUserRoles
func (f FrontendUserRoles) IsValid() bool {
	// check if values are set
	if len(f.Email) == 0 || len(f.Roles) == 0 || !IsValidEmail(f.Email) {
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

// FrontendUserBackendState represents the state of the user in the backend
type FrontendUserBackendState struct {
	UID      string     `json:"uid,omitempty"`
	Email    string     `json:"email,omitempty"`
	Roles    []Role     `json:"roles,omitempty"`
	Created  *time.Time `json:"created,omitempty"`
	Modified *time.Time `json:"modified,omitempty"`
}

func (l FrontendUserBackendState) ToFrontendUserClientState() FrontendUserClientState {
	return FrontendUserClientState{
		UID:   l.UID,
		Email: l.Email,
		Roles: l.Roles,
	}
}
