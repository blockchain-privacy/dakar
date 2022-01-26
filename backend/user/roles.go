package user

import (
	"backend/constants"
	"errors"
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

	// role maps; we need to look up roles very often; string slice lookup are slower than map lookups
	// even for small slices: https://www.golangprograms.com/golang-slice-vs-map-benchmark-testing.html
	adminRoleMap       = map[string]bool{allRoutes: true}
	defaultUserRoleMap = map[string]bool{
		// data
		constants.GetRouteTransaction():        true,
		constants.GetRouteBlock():              true,
		constants.GetRouteAddress():            true,
		constants.GetRouteSearch():             true,
		constants.GetRouteAddressOutputRange(): true,
		// user
		constants.GetRouteModifyUser(): true,
		constants.GetRouteDeleteUser(): true,
	}
	privilegedRoleMap = map[string]bool{
		// data
		constants.GetRouteTransaction():        true,
		constants.GetRouteBlock():              true,
		constants.GetRouteAddress():            true,
		constants.GetRouteMeta():               true,
		constants.GetRouteSearch():             true,
		constants.GetRouteAddressOutputRange(): true,
		// user
		constants.GetRouteModifyUser(): true,
		constants.GetRouteDeleteUser(): true,
		// heuristics
		constants.GetRouteHeuristicStatus():      true,
		constants.GetRouteHeuristicDetails():     true,
		constants.GetRouteHeuristicsExecution():  true,
		constants.GetRouteHeuristics():           true,
		constants.GetRouteHeuristicsSummary():    true,
		constants.GetRouteHeuristicList():        true,
		constants.GetRouteDeleteHeuristic():      true,
		constants.GetRouteHeuristicDescriptors(): true,
		// analytics
		constants.GetRouteShortestTransactionPath(): true,
		constants.GetRouteConnectionLookup():        true,
		constants.GetRouteMixingActivity():          true,
		// clusters
		constants.GetRouteClusterLookup():     true,
		constants.GetRouteClusterSummary():    true,
		constants.GetRouteAddCluster():        true,
		constants.GetRouteDeleteCluster():     true,
		constants.GetRouteDeleteAllClusters(): true,
		constants.GetRouteClusterOverview():   true,
		//constants.GetRouteHMILookup():     true,
		// Attribution
		constants.GetRouteAddAttribution():        true,
		constants.GetRouteAttributionOverview():   true,
		constants.GetRouteDeleteAttribution():     true,
		constants.GetRouteDeleteAllAttributions(): true,
	}

	errorRoleDoesNotExist = errors.New("error role does not exist")
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

// GetRoleByName returns for a given role name the Role object
func GetRoleByName(name string) (Role, error) {
	returnedRole, ok := roleMap[name]
	if !ok {
		return nil, errorRoleDoesNotExist
	}

	return returnedRole, nil
}
