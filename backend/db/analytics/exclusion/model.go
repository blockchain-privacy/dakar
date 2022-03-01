package exclusion

type User struct {
	UID        string              `json:"uid,omitempty"`
	Exclusions []AddressExclusions `json:"User.addressExclusions,omitempty"`
}

type AddressExclusions struct {
	UID string `json:"uid,omitempty"`
}
