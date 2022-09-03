package server

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewAdminRole(t *testing.T) {
	adminRole := NewAdminRole()
	require.NotEmpty(t, adminRole.String(), "string representation of role is empty")
	require.NotEmpty(t, adminRole.GetName(), "name of role is empty")
	require.NotNil(t, adminRole.GetAllowedRoutes())

	for k := range defaultUserRoleMap {
		if !adminRole.IsRouteAllowed(k) {
			t.Fatal(k, "should be allowed for admin")
		}
	}

	for k := range privilegedRoleMap {
		if !adminRole.IsRouteAllowed(k) {
			t.Fatal(k, "should be allowed for admin")
		}
	}
}

func TestNewDefaultUserRole(t *testing.T) {
	userRole := NewDefaultUserRole()

	require.NotEmpty(t, userRole.String(), "string representation of role is empty")
	require.NotEmpty(t, userRole.GetName(), "name of role is empty")
	require.NotNil(t, userRole.GetAllowedRoutes())

	for k := range defaultUserRoleMap {
		if !userRole.IsRouteAllowed(k) {
			t.Fatal(k, "should be allowed for user")
		}
	}

	for k := range privilegedRoleMap {
		if userRole.IsRouteAllowed(k) && !defaultUserRoleMap[k] {
			t.Fatal(k, "should not be allowed for user")
		}
	}
}

func TestNewPrivilegedRole(t *testing.T) {
	privRole := NewPrivilegedRole()

	require.NotEmpty(t, privRole.String(), "string representation of role is empty")
	require.NotEmpty(t, privRole.GetName(), "name of role is empty")
	require.NotNil(t, privRole.GetAllowedRoutes())

	for k := range privilegedRoleMap {
		if !privRole.IsRouteAllowed(k) {
			t.Fatal(k, "should be allowed for privileged user")
		}
	}

	for k := range defaultUserRoleMap {
		if privRole.IsRouteAllowed(k) && !privilegedRoleMap[k] {
			t.Fatal(k, "should not be allowed for privileged user")
		}
	}
}

func TestGetRoleByName(t *testing.T) {
	adminRole, err := getRoleByName(AdminRoleName)
	require.Nil(t, err)
	require.NotNil(t, adminRole)

	userRole, err := getRoleByName("user")
	require.Nil(t, err)
	require.NotNil(t, userRole)

	privRole, err := getRoleByName("privileged")
	require.Nil(t, err)
	require.NotNil(t, privRole)

	invRole, err := getRoleByName("some_invalid_role_string")
	require.NotNil(t, err)
	require.Nil(t, invRole)
}
