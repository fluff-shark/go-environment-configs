
[![BuildStatus](https://travis-ci.com/fluff-shark/go-environment-configs.svg?branch=master)](https://travis-ci.com/fluff-shark/go-environment-configs)
[![ReportCard](https://goreportcard.com/badge/github.com/fluff-shark/go-environment-configs)](https://goreportcard.com/report/github.com/fluff-shark/go-environment-configs)
[![GoDoc](https://godoc.org/github.com/fluff-shark/go-environment-configs?status.svg)](https://godoc.org/github.com/fluff-shark/go-environment-configs)

# Overview

This library helps applications work with configs through environment variables.
It's a more opinionated and much smaller version of
[Viper](https://github.com/spf13/viper). It assumes you're following the
12 factor app recommendations for [configs](https://12factor.net/config) and
[logging](https://12factor.net/logs).

# Goals

Applications generally want to do the following:

1. Define some default config values which work in development.
2. Let apps overwrite defaults with environment variables.
3. Validate the config values.
4. Log values of any config variables _except_ for credentials.

This library's goal is to make these steps as painless as possible.

# Usage

Define structs and tag them with the names of environment variables:

```go
type struct Config {
  Main Server `environment:"MAIN"`
  Admin Server `environment:"ADMIN"`
  Password string `environment:"PASSWORD"`
}

type struct Server {
  Port int `environment:"PORT"`
}
```

Set some environment variables in your shell.

```sh
export MYAPP_MAIN_PORT=80
export MYAPP_ADMIN_PORT=81
export MYAPP_PASSWORD=boo
```

Reference them in the code like so:

```go
import (
  "github.com/fluff-shark/go-environment-configs"
)

func Parse() Config {
  // Define defaults by setting the initial struct values.
  cfg := Config{
    Main: Server{
      Port: 80
    },
    Admin: Server{
      Port: 81
    }
  }

  // Overwrite the defaults with environment variables.
  // Panic with a descriptive error message if the values don't match the types.
  configs.MustLoadWithPrefix(&cfg, "MYAPP")

  // Print the config values.
  // Anything named "password" will be logged as "<redacted>"
  configs.LogWithPrefix(&cfg, "MYAPP")
}
```

The "Prefix" is intended as a namespace to help separate your app's environment
variables from others running on the same system.

This library only handles type-based valiation. The Load() functions will return errors
or panic if an environment variable has a value which can't fit into the Config struct,
but any extra validation is on the caller. For example:

```go
  cfg := Config{
    // Set defaults like above
  }

  err := configs.LoadWithPrefix(&cfg, "MYAPP")

  // Add app-specific validation errors. The library will collect these
  // alongside errors generated by LoadWithPrefix for pretty printing.
  err = configs.Ensure(err, "MYAPP_MAIN_PORT", cfg.Main.Port > 0, "must be a positive integer")
```

# Contributing

Pull requests are welcome for bugfixes and support for types
that aren't yet implemented. Otherwise please open an issue
first to discuss.
