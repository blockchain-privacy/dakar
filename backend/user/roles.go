package user

import (
	"backend/constants"
	"errors"
	"fmt"
)

// allRoutes is the value which should be set if a Role is allowed to use all possible routes
const allRoutes = "ALL_ROUTES"

var (
	// roleMap holds all possible Role mappings
	roleMap = map[string]Role{"admin": NewAdminRole(),
		"user": NewDefaultUserRole(), "privileged": NewPrivilegedRole()}

	// role maps; we need to look up roles very often; string slice lookup are slower than map look ups even for small sices:
	// https://www.golangprograms.com/golang-slice-vs-map-benchmark-testing.html
	adminRoleMap       = map[string]bool{allRoutes: true}
	defaultUserRoleMap = map[string]bool{constants.GetRouteTransaction(): true, constants.GetRouteBlock(): true,
		constants.GetRouteAddress(): true, constants.GetRouteMeta(): true, constants.GetRouteSearch(): true,
		constants.GetRouteAddressOutputRange(): true}
	privilegedRoleMap = map[string]bool{constants.GetRouteTransaction(): true, constants.GetRouteBlock(): true,
		constants.GetRouteAddress(): true, constants.GetRouteMeta(): true, constants.GetRouteSearch(): true,
		constants.GetRouteAddressOutputRange(): true, constants.GetRouteHeuristicStatus(): true,
		constants.GetRouteHeuristicDetails(): true, constants.GetRouteHeuristicsExecution(): true,
		constants.GetRouteHeuristics(): true}

	errorRoleDoesNotExist = errors.New("error role does not exist")
)

type Role interface {
	// getName returns the Role name
	getName() string
	// getAllowedRoutes returns all allowed routes
	getAllowedRoutes() map[string]bool
	// isRouteAllowed returns true if the given route is allowed to be executed
	isRouteAllowed(route string) bool
	// String returns a printable string representation
	String() string
}

type AdminRole struct {
	name          string
	allowedRoutes map[string]bool
}

// NewAdminRole constructor
func NewAdminRole() AdminRole {
	return AdminRole{
		name:          "admin",
		allowedRoutes: adminRoleMap,
	}
}

func (a AdminRole) getName() string {
	return a.name
}

func (a AdminRole) getAllowedRoutes() map[string]bool {
	return a.allowedRoutes
}

func (a AdminRole) isRouteAllowed(_ string) bool {
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

func (a AdminRole) String() string {
	return fmt.Sprintf("Name: %s, Allowed routes: [%s]", a.name, routeMapToString(a.allowedRoutes))
}

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

func (d DefaultUserRole) getName() string {
	return d.name
}

func (d DefaultUserRole) getAllowedRoutes() map[string]bool {
	return d.allowedRoutes
}

func (d DefaultUserRole) isRouteAllowed(route string) bool {
	return d.allowedRoutes[route]
}

func (d DefaultUserRole) String() string {
	return fmt.Sprintf("Name: %s, Allowed routes: [%s]", d.name, routeMapToString(d.allowedRoutes))
}

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

func (p PrivilegedRole) getName() string {
	return p.name
}

func (p PrivilegedRole) getAllowedRoutes() map[string]bool {
	return p.allowedRoutes
}

func (p PrivilegedRole) isRouteAllowed(route string) bool {
	return p.allowedRoutes[route]
}

func (p PrivilegedRole) String() string {
	return fmt.Sprintf("Name: %s, Allowed routes: [%s]", p.name, routeMapToString(p.allowedRoutes))
}

func GetRoleByName(name string) (Role, error) {
	returnedRole, ok := roleMap[name]
	if !ok {
		return nil, errorRoleDoesNotExist
	}

	return returnedRole, nil
}
