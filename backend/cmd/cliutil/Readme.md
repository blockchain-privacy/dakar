# CLI-Util

This module supports reading configuration values from a YAML file.

## Using this module

Import the module, create a struct containing all the needed configuration options and add YAML flags for them.
Call `flag.Parse()` if the command-line options are wanted.

## Example

```go
package main

import (
    flag
	cli "backend/cmd/cliutil"
)

type Config struct {
	SaveDir string `yaml:"saveDir"`
	Logfile string `yaml:"logfile"`
}

var defaultConfig = Config{
	SaveDir: "/some/path/to/a/dir",
	Logfile: "/some/path/to/a/file",
}

func main() {
	var config Config
	if err := cli.GetConfig("config.yml", &config, defaultConfig); err != nil {
		log.Println(err)
		return
	}
    flag.Parse()
    fmt.Println(config.Logfile)
}
```

## Available CLI flags

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| createConfig | false | creates a default config file (default: false) |
| config | < empty string > | config file path (default: < the file name passed to GetConfig >) |
