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

func (r *Role) String() string {
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
	KratosID     string                 `json:"User.kratosID,omitempty"`
	DType        []string               `json:"dgraph.type,omitempty"`
}

func (u *User) String() string {
	return fmt.Sprintf("uid: %s, kratosID: %s, email %s, roles: %v, created: %s, modified: %s, heuristic count: %d",
		u.UID, u.KratosID, u.Email, u.Roles, u.Created, u.Modified, len(u.Heuristics))
}

// SetDType sets the DType for dgraph type recognition
func (u *User) SetDType() {
	u.DType = []string{DTypeUser}
}

// ToFrontendUserStateWithCredentials returns user data for the frontend
func (u *User) ToFrontendUserStateWithCredentials() FrontendUserClientStateWithCredentials {
	roles := make([]FrontendRole, len(u.Roles))

	for i, r := range u.Roles {
		roles[i] = FrontendRole{UID: r.UID, Name: r.Name}
	}

	return FrontendUserClientStateWithCredentials{
		UID:    u.UID,
		Email:  u.Email,
		Roles:  roles,
		Pwhash: u.PasswordHash,
	}
}

// ToFrontendUserBackendState converts frontend user data to the user backend representation
func (u *User) ToFrontendUserBackendState() FrontendUserBackendState {
	roles := make([]FrontendRole, len(u.Roles))

	for i, r := range u.Roles {
		roles[i] = FrontendRole{UID: r.UID, Name: r.Name}
	}

	return FrontendUserBackendState{
		UID:      u.UID,
		Email:    u.Email,
		Roles:    roles,
		Modified: u.Modified,
		Created:  u.Created,
		KratosID: u.KratosID,
	}
}

// FrontendUserClientStateWithCredentials represents the client side state of the user
type FrontendUserClientStateWithCredentials struct {
	UID    string         `json:"uid,omitempty"`
	Email  string         `json:"email,omitempty"`
	Pwhash string         `json:"pwhash,omitempty"`
	Roles  []FrontendRole `json:"roles,omitempty"`
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
	roles := make([]Role, len(f.Roles))

	for i, r := range f.Roles {
		roles[i] = Role{Name: r}
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

// FrontendUserBackendState represents the state of the user in the backend
type FrontendUserBackendState struct {
	UID      string         `json:"uid,omitempty"`
	Email    string         `json:"email,omitempty"`
	KratosID string         `json:"kratosID,omitempty"`
	Roles    []FrontendRole `json:"roles,omitempty"`
	Created  *time.Time     `json:"created,omitempty"`
	Modified *time.Time     `json:"modified,omitempty"`
}
