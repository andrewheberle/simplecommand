package simplecommand_test

import (
	"context"
	"fmt"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
	"github.com/spf13/cobra"
)

func ExampleNew() {
	// Here we create a simple command that does nothing
	command := simplecommand.New("example-command", "This is an example command that does nothing")

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--help"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output:
	// This is an example command that does nothing
	//
	// Usage:
	//   example-command [flags] [args]
	//
	// Flags:
	//   -h, --help   help for example-command
}

func ExampleWithLong() {
	// Here we create a simple command that does nothing
	command := simplecommand.New("example-command",
		"This is an example command that does nothing",
		simplecommand.WithLong(`Here is a much longer help description of this example-command
that is shown when the --help flag is provided.

This can include line breaks and be as long as you like.`),
	)

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--help"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output:
	// Here is a much longer help description of this example-command
	// that is shown when the --help flag is provided.
	//
	// This can include line breaks and be as long as you like.
	//
	// Usage:
	//   example-command [flags] [args]
	//
	// Flags:
	//   -h, --help   help for example-command
}

// this is our custom command type
type ourCommand struct {
	// flags
	exampleFlag string

	// embed the simplecommand.Command type
	*simplecommand.Command
}

// The Init method is implemented to handle our command line flags, however we also run the default *Command.Init method
// to minimise our work a little (ie setting "Short", "Long" and "Deprecated")
func (c *ourCommand) Init(cd *simplecobra.Commandeer) error {
	// run default Init to set up Long/Short/Deprecated
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	cmd := cd.CobraCommand
	cmd.Flags().StringVar(&c.exampleFlag, "example", "", "Example flag")

	return nil
}

// The Run method is implemented to do our actual work
func (c *ourCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	fmt.Printf("Ran \"%s\" with the example flag set to \"%s\"\n", c.Name(), c.exampleFlag)

	return nil
}

func ExampleNew_embedded() {
	// Here we create a simple command using our custom type
	command := &ourCommand{
		Command: simplecommand.New("example-command", "This is an example command"),
	}

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--example", "test"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output: Ran "example-command" with the example flag set to "test"
}

func ExampleNew_subCommand() {
	// Here we create a command that has one sub-command
	rootCommand := simplecommand.New("example-command", "This is an example command that has a single sub-command",
		simplecommand.WithSubCommands(
			&ourCommand{
				Command: simplecommand.New("sub-command", "This is an example sub-command"),
			},
		),
	)

	// Set up simplecobra
	x, err := simplecobra.New(rootCommand)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"sub-command", "--example", "another value"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output: Ran "sub-command" with the example flag set to "another value"
}

func ExampleWithSubCommands() {
	// Here we create a command with two sub-commands, one of which is deprecated
	rootCommand := simplecommand.New("example-command", "This is an example command with sub-commands",
		simplecommand.WithSubCommands(
			simplecommand.New("sub-command", "This is an example sub-command"),

			// this sub-command will not appear in help output
			simplecommand.New("old-command", "This is an old-command", simplecommand.WithDeprecated("use sub-command instead")),
		),
	)

	// Set up simplecobra
	x, err := simplecobra.New(rootCommand)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--help"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output:
	// This is an example command with sub-commands
	//
	// Usage:
	//   example-command [command] [flags]
	//   example-command [command]
	//
	// Available Commands:
	//   completion  Generate the autocompletion script for the specified shell
	//   help        Help about any command
	//   sub-command This is an example sub-command
	//
	// Flags:
	//   -h, --help   help for example-command
	//
	// Use "example-command [command] --help" for more information about a command.
}

func ExampleWithDeprecated() {
	// Here we create a command with a deprecated sub-command
	rootCommand := simplecommand.New("example-command", "This is an example command with a deprecated sub-command",
		simplecommand.WithSubCommands(
			&ourCommand{
				Command: simplecommand.New("old-command", "This is an old-command", simplecommand.WithDeprecated("use sub-command instead")),
			},
		),
	)

	// Set up simplecobra
	x, err := simplecobra.New(rootCommand)
	if err != nil {
		panic(err)
	}

	// a deprecated command still runs, however a message is first written to
	// stderr: Command "old-command" is deprecated, use sub-command instead
	args := []string{"old-command", "--example", "still works"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output: Ran "old-command" with the example flag set to "still works"
}

func ExampleWithAliases() {
	// Here we create a command with a sub-command that can also be run as "sc"
	rootCommand := simplecommand.New("example-command", "This is an example command with an aliased sub-command",
		simplecommand.WithSubCommands(
			&ourCommand{
				Command: simplecommand.New("sub-command", "This is an example sub-command", simplecommand.WithAliases("sc")),
			},
		),
	)

	// Set up simplecobra
	x, err := simplecobra.New(rootCommand)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"sc", "--example", "via alias"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output: Ran "sub-command" with the example flag set to "via alias"
}

func ExampleWithArgs() {
	// Here we create a command that requires exactly one positional argument
	command := simplecommand.New("example-command", "This is an example command that takes one argument",
		simplecommand.WithArgs(cobra.ExactArgs(1)),
	)

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command without an argument, which returns an error
	if _, err := x.Execute(context.Background(), nil); err != nil {
		fmt.Printf("error: %s\n", err)
	}

	// Output: error: command error: accepts 1 arg(s), received 0
}

func ExampleWithExample() {
	// Here we create a command with an example of its use shown in its help
	command := simplecommand.New("example-command", "This is an example command",
		simplecommand.WithExample("  example-command --help"),
	)

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--help"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output:
	// This is an example command
	//
	// Usage:
	//   example-command [flags] [args]
	//
	// Examples:
	//   example-command --help
	//
	// Flags:
	//   -h, --help   help for example-command
}

func ExampleWithHidden() {
	// Here we create a command with a hidden sub-command, which does not appear in help output
	rootCommand := simplecommand.New("example-command", "This is an example command with a hidden sub-command",
		simplecommand.WithSubCommands(
			simplecommand.New("sub-command", "This is an example sub-command"),
			simplecommand.New("debug-command", "This is a hidden sub-command", simplecommand.WithHidden()),
		),
	)

	// Set up simplecobra
	x, err := simplecobra.New(rootCommand)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--help"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output:
	// This is an example command with a hidden sub-command
	//
	// Usage:
	//   example-command [command] [flags]
	//   example-command [command]
	//
	// Available Commands:
	//   completion    Generate the autocompletion script for the specified shell
	//   help          Help about any command
	//   sub-command   This is an example sub-command
	//
	// Flags:
	//   -h, --help   help for example-command
	//
	// Use "example-command [command] --help" for more information about a command.
}

func ExampleWithVersion() {
	// Here we create a command with a version, which adds a --version flag
	command := simplecommand.New("example-command", "This is an example command",
		simplecommand.WithVersion("1.2.3"),
	)

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--version"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output: example-command version 1.2.3
}
