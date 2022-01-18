package attribution

const DType = "Attribution"

type HollowUser struct {
	Uid string `json:"uid,omitempty"`
}

type HollowAddress struct {
	Uid string `json:"uid,omitempty"`
}

type Attribution struct {
	Uid       string        `json:"uid,omitempty"`
	Timestamp string        `json:"attribution_ts,omitempty"`
	Address   HollowAddress `json:"attribution_address,omitempty"`
	Tag       string        `json:"attribution_tag,omitempty"`
	User      HollowUser    `json:"attribution_user,omitempty"`
	DType     []string      `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (a *Attribution) SetDType() {
	a.DType = []string{DType}
}

type FrontendAttribution struct {
	Uid          string   `json:"uid,omitempty"`
	Address      string   `json:"address,omitempty"`
	Tag          string   `json:"tag,omitempty"`
	Timestamp    string   `json:"ts,omitempty"`
	Attributions []string `json:"attributions,omitempty"`
}
