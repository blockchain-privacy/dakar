package server

const (
	// AdminRoleName is the name of the role which is allowed all actions
	AdminRoleName = "admin"
)

// roleMap holds all possible Role mappings
var roleMap = map[string]bool{AdminRoleName: true, "privileged": true}

// isRoleValid returns true if the given role is valid
func isRoleValid(roleName string) bool {
	return roleMap[roleName]
}
