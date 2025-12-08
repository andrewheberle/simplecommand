# simplecommand

[![Go Report Card](https://goreportcard.com/badge/github.com/andrewheberle/simplecommand)](https://goreportcard.com/report/github.com/andrewheberle/simplecommand)
[![GoDoc](https://godoc.org/github.com/andrewheberle/simplecommand?status.svg)](https://godoc.org/github.com/andrewheberle/simplecommand)
[![codecov](https://codecov.io/gh/andrewheberle/simplecommand/graph/badge.svg?token=JEFWB2U0GY)](https://codecov.io/gh/andrewheberle/simplecommand)

This module provides a `*Command` type that satisfies the `simplecobra.Commander` interface.

The main motivation for this module is to only have to implement the bare minimum of methods for any custom commands that are implemented with [github.com/bep/simplecobra](https://github.com/bep/simplecobra).

## Example

As an example, your command may only need to implement a `Run` method as it may not rely on any command-line flags, which would look something like this:

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/andrewheberle/simplecommand"
    "github.com/bep/simplecobra"
)

type subCommand struct {
    *simplecommand.Command
}

func (c *subCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
    fmt.Printf("This is where the work would be done in sub-command \"%s\" for \"%s\"\n", c.Name(), cd.Root.Command.Name())

    return nil
}

func main() {
    rootCmd := simplecommand.New("root-command", "This is an example root-command")
    rootCmd.SubCommands = []simplecobra.Commander{
        &subCommand{
            Command: simplecommand.New("sub-command", "This is an example sub-command"),
        },
    }

    // Set up simplecobra
    x, err := simplecobra.New(rootCmd)
    if err != nil {
        panic(err)
    }

    // run command with the provided args
    if _, err := x.Execute(context.Background(), os.Args[1:]); err != nil {
        panic(err)
    }
}
```

## Viper Integration

[![GoDoc](https://godoc.org/github.com/andrewheberle/simplecommand/vipercommand?status.svg)](https://godoc.org/github.com/andrewheberle/simplecommand/vipercommand)

An alternate implementation of the `simplecobra.Commander` interface is
provided by `*vipercommand.Command`.

This functionality was previously included in `*Command` however this mean that
`viper` and it's associated dependencies were required even if these features
were not used.

To use this functionality you can simple replace `*simpleviper.Command` in your
command's struct with `*vipercommand.Command`.
