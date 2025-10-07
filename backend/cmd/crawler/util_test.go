// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	require.NotPanics(t, func() {
		printVersion("Dash")
	})
}
