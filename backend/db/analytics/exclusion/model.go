// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exclusion

import "backend/db"

type User struct {
	UID        string       `json:"uid,omitempty"`
	Exclusions []db.UIDNode `json:"User.addressExclusions,omitempty"`
}
