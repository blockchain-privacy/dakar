// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package attribution

import "gitlab.com/blockchain-privacy/dakar/db"

const DType = "Attribution"

type Attribution struct {
	UID         string      `json:"uid,omitempty"`
	Timestamp   string      `json:"Attribution.ts,omitempty"`
	Address     *db.UIDNode `json:"Attribution.address,omitempty"`
	Tag         string      `json:"Attribution.tag,omitempty"`
	Description string      `json:"Attribution.description,omitempty"`
	Source      string      `json:"Attribution.source,omitempty"`
	Category    string      `json:"Attribution.category,omitempty"`
	IsPublic    bool        `json:"Attribution.isPublic"`
	User        *db.UIDNode `json:"Attribution.user,omitempty"`
	DType       []string    `json:"dgraph.type,omitempty"`
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
