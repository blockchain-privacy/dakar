package cliutil

import (
	"flag"
	"github.com/qrest/gomisc/serror"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReadConfig reads the config file from the given file path
func ReadConfig(configFilePath string, config interface{}) error {
	// Open config file
	file, err := os.Open(configFilePath)
	if err != nil {
		return serror.NewStackError(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(serror.NewStackError(err))
		}
	}(file)

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	d.KnownFields(true)

	// Start YAML decoding from file
	if err := d.Decode(config); err != nil {
		return serror.NewStackError(err)
	}

	return nil
}

// WriteConfig writes the config value to the given file path
func WriteConfig(filePath string, config interface{}) error {
	marshalledConfig, err := yaml.Marshal(&config)
	if err != nil {
		return serror.NewStackError(err)
	}

	if err := os.WriteFile(filePath, marshalledConfig, 0600); err != nil {
		return serror.NewStackError(err)
	}

	return nil
}

// SetConfigFlags sets the CLI flags for accessing and generating the configuration file
func SetConfigFlags(defaultConfigName string, filePath *string, createConfigFile *bool) {
	flag.StringVar(filePath, "config", defaultConfigName, "config file path")
	flag.BoolVar(createConfigFile, "createConfig", false,
		"creates a default config file '"+defaultConfigName+"' (default: false)")
}

// BuildEndpoint creates a string in the format of "host:port"
func BuildEndpoint(host string, port uint) (string, error) {
	host = strings.TrimSpace(host)
	if len(host) == 0 || port == 0 {
		return "", serror.NewStackErrorStr("host or port is not valid")
	}

	return host + ":" + strconv.Itoa(int(port)), nil
}

// GetLogfile returns a file accessor for fileName
func GetLogfile(fileName string) (f *os.File, err error) {
	if len(fileName) == 0 {
		err = serror.NewStackErrorStr("name for log file is invalid")
		return
	}

	f, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		err = serror.NewStackError(err)
		return
	}

	return
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
