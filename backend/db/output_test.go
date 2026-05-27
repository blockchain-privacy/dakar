// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutput_SetDType(t *testing.T) {
	output := Output{}
	output.SetDType()
	require.Equal(t, []string{outputDType}, output.DType)
}
