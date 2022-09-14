package server

import (
	dbus "backend/db/user"

	"time"
)

const (
	// tokenExpirationTime is the time a token is valid
	tokenExpirationTime = time.Hour * 24
	// if only reissueDuration is left of the token lifetime it gets reissued
	reissueDuration = tokenExpirationTime / 4
)

type tokenUser struct {
	ID       string      `json:"uid,omitempty"`
	KratosID string      `json:"kratos_id,omitempty"`
	Roles    []dbus.Role `json:"roles,omitempty"`
}
