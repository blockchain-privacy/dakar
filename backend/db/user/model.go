package user

import (
	"backend/db/analytics/heuristics"
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
	Name  string   `json:"Role.name"`
	DType []string `json:"dgraph.type,omitempty"`
}

func (r Role) String() string {
	return fmt.Sprintf("uid: %s, name: %s", r.UID, r.Name)
}

// SetDType sets the DType for dgraph type recognition
func (r *Role) SetDType() {
	r.DType = []string{DTypeRole}
}

// FrontendRole should be used for client facing responses
type FrontendRole struct {
	UID  string `json:"uid,omitempty"`
	Name string `json:"name"`
}

// User is the database representation of a user
type User struct {
	UID          string                 `json:"uid,omitempty"`
	Email        string                 `json:"User.email,omitempty"`
	PasswordHash string                 `json:"User.pwhash,omitempty"`
	Roles        []Role                 `json:"User.roles,omitempty"`
	Created      *time.Time             `json:"User.created,omitempty"`
	Modified     *time.Time             `json:"User.modified,omitempty"`
	Heuristics   []heuristics.Heuristic `json:"User.heuristics,omitempty"`
	DType        []string               `json:"dgraph.type,omitempty"`
}

func (u User) String() string {
	return fmt.Sprintf("uid: %s, email %s, roles: %v, created: %s, modified: %s, heuristic count: %d",
		u.UID, u.Email, u.Roles, u.Created, u.Modified, len(u.Heuristics))
}

// SetDType sets the DType for dgraph type recognition
func (u *User) SetDType() {
	u.DType = []string{DTypeUser}
}

// ToFrontendUserState returns user data for the frontend
func (u User) ToFrontendUserState() FrontendUserClientState {
	var roles []FrontendRole

	for _, r := range u.Roles {
		convertedRole := FrontendRole{UID: r.UID, Name: r.Name}
		roles = append(roles, convertedRole)
	}

	return FrontendUserClientState{
		UID:   u.UID,
		Email: u.Email,
		Roles: roles,
	}
}

// ToFrontendUserBackendState converts frontend user data to the user backend representation
func (u User) ToFrontendUserBackendState() FrontendUserBackendState {
	var roles []FrontendRole

	for _, r := range u.Roles {
		convertedRole := FrontendRole{UID: r.UID, Name: r.Name}
		roles = append(roles, convertedRole)
	}

	return FrontendUserBackendState{
		UID:      u.UID,
		Email:    u.Email,
		Roles:    roles,
		Modified: u.Modified,
		Created:  u.Created,
	}
}

// ModifyUserRequest holds the configuration data for a user modification request
type ModifyUserRequest struct {
	UID             string         `json:"uid,omitempty"`
	Email           string         `json:"email,omitempty"`
	CurrentPassword string         `json:"current_password,omitempty"`
	NewPassword     string         `json:"new_password,omitempty"`
	Roles           []FrontendRole `json:"roles,omitempty"`
}

// ToUser returns a User object with the given password hash
func (m ModifyUserRequest) ToUser(pwHash string) User {
	var roles []Role

	for _, r := range m.Roles {
		convertedRole := Role{UID: r.UID, Name: r.Name}
		convertedRole.SetDType()
		roles = append(roles, convertedRole)
	}

	return User{
		UID:          m.UID,
		Email:        m.Email,
		Roles:        roles,
		PasswordHash: pwHash,
	}
}

// FrontendUserClientState represents the client side state of the user
type FrontendUserClientState struct {
	UID   string         `json:"uid,omitempty"`
	Email string         `json:"email,omitempty"`
	Roles []FrontendRole `json:"roles,omitempty"`
}

// ToUser returns a User object
func (f FrontendUserClientState) ToUser() User {
	var roles []Role

	for _, r := range f.Roles {
		convertedRole := Role{UID: r.UID, Name: r.Name}
		convertedRole.SetDType()
		roles = append(roles, convertedRole)
	}

	return User{
		UID:   f.UID,
		Email: f.Email,
		Roles: roles,
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

// FrontendUserRoles is the role representation for the frontend
type FrontendUserRoles struct {
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

func (f FrontendUserRoles) String() string {
	return fmt.Sprintf("email %s, roles: %v", f.Email, f.Roles)
}

// ToUser returns a User object
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

// FrontendUserLogin holds data of a user login
type FrontendUserLogin struct {
	Email    string `json:"email"`
	Password string `json:"pw"`
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
	UID      string         `json:"uid,omitempty"`
	Email    string         `json:"email,omitempty"`
	Roles    []FrontendRole `json:"roles,omitempty"`
	Created  *time.Time     `json:"created,omitempty"`
	Modified *time.Time     `json:"modified,omitempty"`
}

// ToFrontendUserClientState converts the backend user state to a frontend user state
func (l FrontendUserBackendState) ToFrontendUserClientState() FrontendUserClientState {
	return FrontendUserClientState{
		UID:   l.UID,
		Email: l.Email,
		Roles: l.Roles,
	}
}
