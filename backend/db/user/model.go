package user

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
	UID   string   `json:"uid,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (u *User) SetDType() {
	u.DType = []string{DTypeUser}
}
