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
    var filePath string
    var createConfigFile bool
    cli.SetConfigFlags(defaultConfigName, &filePath, &createConfigFile)
    flag.Parse()

    if createConfigFile {
        fmt.Println("Generating configuration file ...")

        err := cli.WriteConfig(defaultConfigName, defaultConfig)
        if err != nil {
            fmt.Println(err)
            return
        }

        fmt.Println("config file", defaultConfigName, "successfully created")
        return
    }

    var config Config
    if err := cli.ReadConfig(filePath, &config); err != nil {
        fmt.Println(err)
        return
    }
}
```

## Available CLI flags

| Flag         |  Default Value   |                                                       Description |
|--------------|:----------------:|------------------------------------------------------------------:|
| createConfig |      false       |                    creates a default config file (default: false) |
| config       | < empty string > | config file path (default: < the file name passed to GetConfig >) |
