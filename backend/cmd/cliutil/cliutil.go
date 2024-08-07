package cliutil

import (
	"github.com/qrest/gomisc/serror"
	"strconv"
	"strings"
)

// BuildEndpoint creates a string in the format of "host:port"
func BuildEndpoint(host string, port uint) (string, error) {
	host = strings.TrimSpace(host)
	if len(host) == 0 || port == 0 {
		return "", serror.FromStr("host or port is not valid")
	}

	return host + ":" + strconv.Itoa(int(port)), nil
}

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
// todo: review once go 1.23 is released. Could be replaced with maps.Keys(someMap)
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
// todo: review once go 1.23 is released. Could be replaced with maps.Keys(someMap)
func GetMapValues[M ~map[K]V, K comparable, V any](m M) []V {
	values := make([]V, len(m))
	i := 0
	for _, v := range m {
		values[i] = v
		i++
	}
	return values
}
