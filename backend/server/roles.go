package server

import (
	"backend/cmd/cliutil"
	"fmt"
)

const (
	// allRoutes is the value which should be set if a Role is allowed to use all possible routes
	allRoutes = "ALL_ROUTES"
	// AdminRoleName is the name of the role which is allowed all actions
	AdminRoleName = "admin"
)

var (
	// roleMap holds all possible Role mappings
	roleMap = map[string]Role{AdminRoleName: NewAdminRole(),
		"user": NewDefaultUserRole(), "privileged": NewPrivilegedRole()}

	// role maps; we need to look up roles very often; string slice lookups are slower than map lookups
	// even for small slices: https://boltandnuts.wordpress.com/2017/11/20/go-slice-vs-maps/
	adminRoleMap       = map[string]bool{allRoutes: true}
	defaultUserRoleMap = map[string]bool{
		// data
		getRouteTransaction():        true,
		getRouteBlock():              true,
		getRouteAddress():            true,
		getRouteSearch():             true,
		getRouteAddressOutputRange(): true,
		// user
		getRouteDeleteIdentity(): true,
	}
	privilegedRoleMap = map[string]bool{
		// data
		getRouteTransaction():        true,
		getRouteBlock():              true,
		getRouteAddress():            true,
		getRouteMeta():               true,
		getRouteSearch():             true,
		getRouteAddressOutputRange(): true,
		// user
		getRouteDeleteIdentity(): true,
		// heuristics
		getRouteHeuristicDetails():    true,
		getRouteHeuristicsExecution(): true,
		getRouteHeuristicByWorkID():   true,
		getRouteHeuristicReport():     true,
		// analytics
		getRouteShortestTransactionPath(): true,
		getRouteConnectionLookup():        true,
		getRouteMixingActivity():          true,
		getRouteSpendingFingerprint():     true,
		// clusters
		getRouteClusterLookup():     true,
		getRouteClusterReport():     true,
		getRouteAddCluster():        true,
		getRouteDeleteCluster():     true,
		getRouteDeleteAllClusters(): true,
		getRouteClusterOverview():   true,
		// getRouteHMILookup():     true,
		// Attribution
		getRouteAddPrivateAttribution():        true,
		getRouteAttributionList():              true,
		getRouteDeletePrivateAttribution():     true,
		getRouteDeleteAllPrivateAttributions(): true,
		getRouteSearchAttributions():           true,
		// Address exclusion
		getRouteAddressExclusionList():       true,
		getRouteAddressExclusionStatus():     true,
		getRouteDeleteAddressExclusion():     true,
		getRouteAddAddressExclusions():       true,
		getRouteDeleteAllAddressExclusions(): true,
		// workspace
		getRouteWorkspaceAddNode():    true,
		getRouteWorkspaces():          true,
		getRouteAddWorkspace():        true,
		getRouteGetWorkspace():        true,
		getRouteUpdateWorkspace():     true,
		getRouteDeleteWorkspace():     true,
		getRouteDeleteAllWorkspaces(): true,
		getRouteWorkspaceDeleteNode(): true,
	}
)

// Role defines an interface which allows to access the properties of Roles
type Role interface {
	// GetName returns the Role name
	GetName() string
	// GetAllowedRoutes returns all allowed routes
	GetAllowedRoutes() map[string]bool
	// IsRouteAllowed returns true if the given route is allowed to be executed
	IsRouteAllowed(route string) bool
	// String returns a printable string representation
	String() string
}

// AdminRole has access to all possible routes
type AdminRole struct {
	name          string
	allowedRoutes map[string]bool
}

// NewAdminRole constructor
func NewAdminRole() AdminRole {
	return AdminRole{
		name:          AdminRoleName,
		allowedRoutes: adminRoleMap,
	}
}

// GetName returns the name of the role
func (a AdminRole) GetName() string {
	return a.name
}

// GetAllowedRoutes returns all routes which this role is allowed to access
func (a AdminRole) GetAllowedRoutes() map[string]bool {
	return a.allowedRoutes
}

// IsRouteAllowed returns true if the role is allowed to access the given route
func (a AdminRole) IsRouteAllowed(_ string) bool {
	return true
}

func routeMapToString(m map[string]bool) string {
	var allowedRoutesString string
	i := 0
	for k := range m {
		allowedRoutesString += k

		if i < len(m) {
			allowedRoutesString += " "
		}

		i++
	}
	return allowedRoutesString
}

// String returns the string representation of the role
func (a AdminRole) String() string {
	return fmt.Sprintf("Name: %s, Allowed routes: [%s]", a.name, routeMapToString(a.allowedRoutes))
}

// DefaultUserRole has access to basic routes
type DefaultUserRole struct {
	name          string
	allowedRoutes map[string]bool
}

// NewDefaultUserRole constructor
func NewDefaultUserRole() DefaultUserRole {
	return DefaultUserRole{
		name:          "user",
		allowedRoutes: defaultUserRoleMap,
	}
}

// GetName returns the name of the role
func (d DefaultUserRole) GetName() string {
	return d.name
}

// GetAllowedRoutes returns all routes which this role is allowed to access
func (d DefaultUserRole) GetAllowedRoutes() map[string]bool {
	return d.allowedRoutes
}

// IsRouteAllowed returns true if the role is allowed to access the given route
func (d DefaultUserRole) IsRouteAllowed(route string) bool {
	return d.allowedRoutes[route]
}

// String returns the string representation of the role
func (d DefaultUserRole) String() string {
	return fmt.Sprintf("Name: %s, Allowed routes: [%s]", d.name, routeMapToString(d.allowedRoutes))
}

// PrivilegedRole has access to basic routes and routes which perform computationally heavy tasks
type PrivilegedRole struct {
	name          string
	allowedRoutes map[string]bool
}

// NewPrivilegedRole constructor
func NewPrivilegedRole() PrivilegedRole {
	return PrivilegedRole{
		name:          "privileged",
		allowedRoutes: privilegedRoleMap,
	}
}

// GetName returns the name of the role
func (p PrivilegedRole) GetName() string {
	return p.name
}

// GetAllowedRoutes returns all routes which this role is allowed to access
func (p PrivilegedRole) GetAllowedRoutes() map[string]bool {
	return p.allowedRoutes
}

// IsRouteAllowed returns true if the role is allowed to access the given route
func (p PrivilegedRole) IsRouteAllowed(route string) bool {
	return p.allowedRoutes[route]
}

// String returns the string representation of the role
func (p PrivilegedRole) String() string {
	return fmt.Sprintf("Name: %s, Allowed routes: [%s]", p.name, routeMapToString(p.allowedRoutes))
}

// getRoleByName returns for a given role name the Role object
func getRoleByName(name string) (Role, error) {
	returnedRole, ok := roleMap[name]
	if !ok {
		return nil, cliutil.NewStackErrorStr("role does not exist")
	}

	return returnedRole, nil
}
