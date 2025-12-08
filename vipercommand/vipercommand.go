package vipercommand

import (
	"strings"

	"github.com/andrewheberle/simplecommand"
	"github.com/andrewheberle/simpleviper"
	"github.com/bep/simplecobra"
	"github.com/spf13/viper"
)

type Command struct {
	// Config specifies a configuration file used
	Config string

	// Allow missing config file when Config is set
	ConfigOptional bool

	// Enviroment variable handling with Viper. See [viper.SetEnvPrefix] for details.
	EnvPrefix string

	// Enviroment variable handling with Viper. See [viper.SetEnvKeyReplacer] for details.
	EnvKeyReplacer *strings.Replacer

	viperlet *simpleviper.Viperlet

	*simplecommand.Command
}

// New creates a bare minimum [*Command] with a name and a short description set
func New(name, short string, opts ...simplecommand.CommandOption) *Command {
	c := &simplecommand.Command{
		CommandName: name,
		Short:       short,
	}

	// set options
	for _, o := range opts {
		o(c)
	}

	return &Command{
		Command: c,
	}
}

// Viper allows access to the underlying [*viper.Viper] instance when enabled.
// Warning: This will return nil if Viper is not enabled for this command.
func (c *Command) Viper() *viper.Viper {
	return c.viperlet.Viper()
}

func (c *Command) PreRun(this, runner *simplecobra.Commandeer) error {
	opts := []simpleviper.Option{
		simpleviper.WithEnvPrefix(c.EnvPrefix),
		simpleviper.WithEnvKeyReplacer(c.EnvKeyReplacer),
	}

	// add config file if set
	if c.Config != "" {
		if c.ConfigOptional {
			opts = append(opts, simpleviper.WithOptionalConfig(c.Config))
		} else {
			opts = append(opts, simpleviper.WithConfig(c.Config))
		}
	}

	// set up viperlet
	c.viperlet = simpleviper.New(opts...)

	// bring in env vars and bind to flagset
	if err := c.viperlet.Init(this.CobraCommand.Flags()); err != nil {
		return err
	}

	return nil
}
