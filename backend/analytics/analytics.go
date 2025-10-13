// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"errors"
)

var (
	ErrTooManyAddresses   = errors.New("request contains too many addresses")
	ErrNonExistentAddress = errors.New("address does not exist")
)
