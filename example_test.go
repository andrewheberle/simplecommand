package simplecommand_test

import (
	"context"
	"fmt"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
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

// this is our custom command type
type ourCommand struct {
	// flags
	exampleFlag string

	// embed the simplecommand.Command type
	*simplecommand.Command
}

func (c *ourCommand) Init(cd *simplecobra.Commandeer) error {
	cmd := cd.CobraCommand
	cmd.Short = c.Short

	cmd.Flags().StringVar(&c.exampleFlag, "example", "", "Example flag")

	return nil
}

func (c *ourCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	fmt.Printf("Ran \"%s\" with the example flag set to \"%s\"\n", cd.CobraCommand.Name(), c.exampleFlag)

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
	rootCommand := simplecommand.New("example-command", "This is an example command that has a single sub-command")
	rootCommand.SubCommands = []simplecobra.Commander{
		&ourCommand{
			Command: simplecommand.New("sub-command", "This is an example sub-command"),
		},
	}

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
