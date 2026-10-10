package koanfcommand_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/andrewheberle/simplecommand/koanfcommand"
	"github.com/bep/simplecobra"
)

type koanfCommand struct {
	// flags
	exampleFlag     string
	exampleLongFlag string

	// embed the *koanfcommand.Command type
	*koanfcommand.Command
}

// The Init method is implemented to handle our command line flags, however we also run the default *Command.Init method
// to minimise our work a little (ie setting "Short", "Long" and "Deprecated")
func (c *koanfCommand) Init(cd *simplecobra.Commandeer) error {
	// run default Init
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	// set up command line flags
	cmd := cd.CobraCommand
	cmd.Flags().StringVar(&c.exampleFlag, "example", "", "Example flag")
	cmd.Flags().StringVar(&c.exampleLongFlag, "example.flag", "", "Example long flag")

	return nil
}

func (c *koanfCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	// Run the default *Command.PreRun method to set flags from the environment and configuration file
	if err := c.Command.PreRun(this, runner); err != nil {
		return err
	}

	// In a real program, not an example, this would be where you would initialise any state
	// required for the command.

	return nil
}

// The Run method is implemented to do our actual work
func (c *koanfCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	fmt.Printf("Ran \"%s\" with the example flag set to \"%s\"\n", c.Name(), c.exampleFlag)
	fmt.Printf("Ran \"%s\" with the example long flag set to \"%s\"\n", c.Name(), c.exampleLongFlag)

	return nil
}

func ExampleNew() {
	// Here we create a simple command using our custom type with koanf enabled
	command := &koanfCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	command.EnvPrefix = "cmd"

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// set our env var
	if err := os.Setenv("CMD_EXAMPLE", "from env var"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Unsetenv("CMD_EXAMPLE"); err != nil {
			panic(err)
		}
	}()

	// run our command with no arguments so our example flag is set from the environment
	if _, err := x.Execute(context.Background(), nil); err != nil {
		panic(err)
	}

	// Output:
	// Ran "example-command" with the example flag set to "from env var"
	// Ran "example-command" with the example long flag set to ""
}

func ExampleNew_withConfig() {
	// Here we create a simple command using our custom type with koanf enabled
	command := &koanfCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	// set our config file
	command.Config = "testconfig.yml"

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our command with no arguments so our example flag is set from the configuration file
	if _, err := x.Execute(context.Background(), nil); err != nil {
		panic(err)
	}

	// Output:
	// Ran "example-command" with the example flag set to "from config file"
	// Ran "example-command" with the example long flag set to ""
}

func ExampleNew_withEnvKeyReplacer() {
	// Here we create a simple command using our custom type with koanf enabled
	command := &koanfCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	command.EnvPrefix = "cmd"
	command.EnvKeyReplacer = strings.NewReplacer(".", "_")

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// set our env var for the long flag
	if err := os.Setenv("CMD_EXAMPLE_FLAG", "from env var"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Unsetenv("CMD_EXAMPLE_FLAG"); err != nil {
			panic(err)
		}
	}()

	// run our command with no arguments so our long flag is set from the environment
	if _, err := x.Execute(context.Background(), nil); err != nil {
		panic(err)
	}

	// Output:
	// Ran "example-command" with the example flag set to ""
	// Ran "example-command" with the example long flag set to "from env var"
}

type configFlagCommand struct {
	exampleFlag string

	*koanfcommand.Command
}

// As the configuration file is read in PreRun, once command line flags have been parsed, the Config field can be
// set from a command line flag
func (c *configFlagCommand) Init(cd *simplecobra.Commandeer) error {
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	cmd := cd.CobraCommand
	cmd.Flags().StringVar(&c.Config, "config", "", "Configuration file")
	cmd.Flags().StringVar(&c.exampleFlag, "example", "", "Example flag")

	return nil
}

func (c *configFlagCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	fmt.Printf("Ran \"%s\" with the example flag set to \"%s\"\n", c.Name(), c.exampleFlag)

	return nil
}

func ExampleNew_configFlag() {
	// Here we create a command that takes the path to its configuration file from the --config flag
	command := &configFlagCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}

	// Set up simplecobra
	x, err := simplecobra.New(command)
	if err != nil {
		panic(err)
	}

	// run our simplecobra command with the provided args, in a real program args would be os.Args[1:]
	args := []string{"--config", "testconfig.yml"}
	if _, err := x.Execute(context.Background(), args); err != nil {
		panic(err)
	}

	// Output: Ran "example-command" with the example flag set to "from config file"
}
