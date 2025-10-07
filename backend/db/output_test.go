// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOutput_String(t *testing.T) {
	output := Output{}
	require.NotEmpty(t, output.String())
	output.Amount = getNumPointer[int64](500)
	require.NotEmpty(t, output.String())
}

func TestOutput_SetDType(t *testing.T) {
	output := Output{}
	output.SetDType()
	require.Equal(t, []string{outputDType}, output.DType)
}
