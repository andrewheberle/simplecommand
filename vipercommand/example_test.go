package vipercommand_test

import (
	"context"
	"fmt"
	"os"

	"github.com/andrewheberle/simplecommand/vipercommand"
	"github.com/bep/simplecobra"
)

type viperCommand struct {
	// flags
	exampleFlag string

	// embed the *vipercommand.Command type
	*vipercommand.Command
}

// The Init method is implemented to handle our command line flags, however we also run the default *Command.Init method
// to minimise our work a little (ie setting "Short", "Long" and "Deprecated")
func (c *viperCommand) Init(cd *simplecobra.Commandeer) error {
	// run default Init to set up Long/Short/Deprecated
	c.Command.Init(cd)

	// set up command line flags
	cmd := cd.CobraCommand
	cmd.Flags().StringVar(&c.exampleFlag, "example", "", "Example flag")

	return nil
}

func (c *viperCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	// Run the default *Command.PreRun method to set up Viper
	if err := c.Command.PreRun(this, runner); err != nil {
		return err
	}

	// In a real program, not an example, this would be where you would initialise any state
	// required for the command.

	return nil
}

// The Run method is implemented to do our actual work
func (c *viperCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	fmt.Printf("Ran \"%s\" with the example flag set to \"%s\"\n", c.Name(), c.exampleFlag)

	return nil
}

func ExampleNew() {
	// Here we create a simple command using our custom type with Viper enabled
	command := &viperCommand{
		Command: vipercommand.New("example-command", "This is an example command (with fangs!)"),
	}
	// set the EnvPrefix
	command.EnvPrefix = "cmd"

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// set our env var
	os.Setenv("CMD_EXAMPLE", "from env var")

	// run our command with no arguments so our example flag is set from the environment, in a real program args would be os.Args[1:]
	if _, err := x.Execute(context.Background(), []string{}); err != nil {
		panic(err)
	}

	// Output: Ran "example-command" with the example flag set to "from env var"
}

func ExampleNew_withconfig() {
	// Here we create a simple command using our custom type with Viper enabled
	command := &viperCommand{
		Command: vipercommand.New("example-command", "This is an example command (with fangs!)"),
	}
	// set our config file
	command.Config = "testconfig.yml"

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our command with no arguments so our example flag is set from the configuration file, in a real program args would be os.Args[1:]
	if _, err := x.Execute(context.Background(), []string{}); err != nil {
		panic(err)
	}

	// Output: Ran "example-command" with the example flag set to "from config file"
}
