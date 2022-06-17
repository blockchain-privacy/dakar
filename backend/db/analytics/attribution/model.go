package attribution

const DType = "Attribution"

type HollowUser struct {
	UID string `json:"uid,omitempty"`
}

type HollowAddress struct {
	UID string `json:"uid,omitempty"`
}

type Attribution struct {
	UID         string         `json:"uid,omitempty"`
	Timestamp   string         `json:"Attribution.ts,omitempty"`
	Address     *HollowAddress `json:"Attribution.address,omitempty"`
	Tag         string         `json:"Attribution.tag,omitempty"`
	Description string         `json:"Attribution.description,omitempty"`
	Source      string         `json:"Attribution.source,omitempty"`
	Category    string         `json:"Attribution.category,omitempty"`
	IsPublic    bool           `json:"Attribution.isPublic"`
	User        *HollowUser    `json:"Attribution.user,omitempty"`
	DType       []string       `json:"dgraph.type,omitempty"`
}

// SetDType sets the DType for dgraph type recognition
func (a *Attribution) SetDType() {
	a.DType = []string{DType}
}

type FrontendAttribution struct {
	UID         string `json:"uid,omitempty"`
	Address     string `json:"address,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Timestamp   string `json:"ts,omitempty"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
	Category    string `json:"category,omitempty"`
	IsPublic    bool   `json:"isPublic"`
}

type RequestAttribution struct {
	UID         string `json:"uid,omitempty"`
	Timestamp   string `json:"Attribution.ts,omitempty"`
	Tag         string `json:"Attribution.tag,omitempty"`
	Description string `json:"Attribution.description,omitempty"`
	Source      string `json:"Attribution.source,omitempty"`
	Category    string `json:"Attribution.category,omitempty"`
	IsPublic    bool   `json:"Attribution.isPublic"`
	Address     struct {
		Hash string `json:"addresshash,omitempty"`
	} `json:"Attribution.address,omitempty"`
}

func (r RequestAttribution) toFrontendAttribution() FrontendAttribution {
	return FrontendAttribution{
		UID:         r.UID,
		Timestamp:   r.Timestamp,
		Address:     r.Address.Hash,
		Tag:         r.Tag,
		Description: r.Description,
		Source:      r.Source,
		Category:    r.Category,
		IsPublic:    r.IsPublic,
	}
}
