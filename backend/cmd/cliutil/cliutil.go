package cliutil

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// readConfig reads the config file from the given file path
func readConfig(configFilePath string, config interface{}) error {
	// Open config file
	file, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("%s: %w", ShowCallInfo(), err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(fmt.Errorf("%s: %w", ShowCallInfo(), err))
		}
	}(file)

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	d.KnownFields(true)

	// Start YAML decoding from file
	if err := d.Decode(config); err != nil {
		return fmt.Errorf("%s: %w", ShowCallInfo(), err)
	}

	return nil
}

func writeConfig(filePath string, config interface{}) error {
	marshalledConfig, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("%s: %w", ShowCallInfo(), err)
	}

	if err := os.WriteFile(filePath, marshalledConfig, 0666); err != nil {
		return fmt.Errorf("%s: %w", ShowCallInfo(), err)
	}

	return nil
}

func setConfigFlags(defaultConfigName string, filePath *string, createConfigFile *bool) {
	flag.NewFlagSet("test", flag.ExitOnError)

	flag.StringVar(filePath, "config", defaultConfigName,
		"config file path (default:"+defaultConfigName+")")
	flag.BoolVar(createConfigFile, "createConfig", false,
		"creates a default config file '"+defaultConfigName+"' (default: false)")
}

func GetConfig(defaultConfigName string, config interface{}, defaultInterface interface{}) error {
	var filePath string
	var createConfigFile bool
	setConfigFlags(defaultConfigName, &filePath, &createConfigFile)
	flag.Parse()
	// create a new config file
	if createConfigFile {
		err := writeConfig(defaultConfigName, defaultInterface)
		if err != nil {
			return fmt.Errorf("%s: %w", ShowCallInfo(), err)
		}

		fmt.Println("config file", defaultConfigName, "successfully created")

		os.Exit(0)
	}

	if err := readConfig(filePath, config); err != nil {
		return fmt.Errorf("%s: %w", ShowCallInfo(), err)
	}

	return nil
}

// ShowCallInfo returns the current call stack
func ShowCallInfo() string {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("not ok")
	}

	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
	}

	return fmt.Sprintf("%s:%d %s", fileName, line, funcName)
}

// BuildEndpoint creates a string in the format of "host:port"
func BuildEndpoint(host string, port uint) (string, error) {
	host = strings.TrimSpace(host)
	if len(host) == 0 || port == 0 {
		return "", errors.New("host or port is not valid")
	}

	return host + ":" + strconv.Itoa(int(port)), nil
}

// GetLogfile returns a file accessor for fileName
func GetLogfile(fileName string) (f *os.File, err error) {
	if len(fileName) == 0 {
		err = errors.New("name for log file is invalid")
		return
	}

	f, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		err = fmt.Errorf("%s: %w", ShowCallInfo(), err)
		return
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	return
}
