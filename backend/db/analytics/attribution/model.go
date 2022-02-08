package attribution

const DType = "Attribution"

type HollowUser struct {
	Uid string `json:"uid,omitempty"`
}

type HollowAddress struct {
	Uid string `json:"uid,omitempty"`
}

type Attribution struct {
	Uid         string         `json:"uid,omitempty"`
	Timestamp   string         `json:"attribution_ts,omitempty"`
	Address     *HollowAddress `json:"attribution_address,omitempty"`
	Tag         string         `json:"attribution_tag,omitempty"`
	Description string         `json:"attribution_description,omitempty"`
	Source      string         `json:"attribution_source,omitempty"`
	Category    string         `json:"attribution_category,omitempty"`
	IsPublic    bool           `json:"attribution_ispublic"`
	User        *HollowUser    `json:"attribution_user,omitempty"`
	DType       []string       `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (a *Attribution) SetDType() {
	a.DType = []string{DType}
}

type FrontendAttribution struct {
	Uid         string `json:"uid,omitempty"`
	Address     string `json:"address,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Timestamp   string `json:"ts,omitempty"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
	Category    string `json:"category,omitempty"`
	IsPublic    bool   `json:"ispublic"`
}

type RequestAttribution struct {
	Uid         string `json:"uid,omitempty"`
	Timestamp   string `json:"attribution_ts,omitempty"`
	Tag         string `json:"attribution_tag,omitempty"`
	Description string `json:"attribution_description,omitempty"`
	Source      string `json:"attribution_source,omitempty"`
	Category    string `json:"attribution_category,omitempty"`
	IsPublic    bool   `json:"attribution_ispublic"`
	Address     struct {
		Hash string `json:"addresshash,omitempty"`
	} `json:"attribution_address,omitempty"`
}

func (r RequestAttribution) toFrontendAttribution() FrontendAttribution {
	return FrontendAttribution{
		Uid:         r.Uid,
		Timestamp:   r.Timestamp,
		Address:     r.Address.Hash,
		Tag:         r.Tag,
		Description: r.Description,
		Source:      r.Source,
		Category:    r.Category,
		IsPublic:    r.IsPublic,
	}
}
