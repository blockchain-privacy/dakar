package exclusion

import "backend/db"

type User struct {
	UID        string       `json:"uid,omitempty"`
	Exclusions []db.UIDNode `json:"User.addressExclusions,omitempty"`
}
