// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cliutil

// GetOneKey returns an indeterminate key of the given map. If the map is empty, an empty key will be returned
func GetOneKey[M ~map[K]V, K comparable, V any](m M) K {
	for k := range m {
		return k
	}

	var key K
	return key
}

// GetOneItem returns an indeterminate key-value pair of the given map.
// If the map is empty, an empty key-value pair will be returned
func GetOneItem[M ~map[K]V, K comparable, V any](m M) (K, V) {
	for k, v := range m {
		return k, v
	}

	var key K
	var val V
	return key, val
}

// GetMapKeys returns all keys of the given map in indeterminate order.
func GetMapKeys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// GetMapValues returns all values of the given map in indeterminate order.
func GetMapValues[M ~map[K]V, K comparable, V any](m M) []V {
	values := make([]V, len(m))
	i := 0
	for _, v := range m {
		values[i] = v
		i++
	}
	return values
}
