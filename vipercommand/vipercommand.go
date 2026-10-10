// Package vipercommand provides a [Command] type that, like
// [simplecommand.Command], satisfies the [simplecobra.Commander] interface but
// also allows command line flags to be set from environment variables and a
// configuration file via [viper].
//
// Command line flags take precedence over environment variables, which take
// precedence over values from the configuration file.
package vipercommand

import (
	"strings"

	"github.com/andrewheberle/simplecommand"
	"github.com/andrewheberle/simpleviper"
	"github.com/bep/simplecobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Command is the basis for creating your own [simplecobra.Commander] with
// [viper] support and is best used by embedding it in your own struct.
//
// A [Command] should always be created using [New], as the embedded
// [*simplecommand.Command] must not be nil.
type Command struct {
	// Config specifies a configuration file used
	Config string

	// Allow missing config file when Config is set
	ConfigOptional bool

	// Environment variable handling with Viper. See [viper.SetEnvPrefix] for
	// details
	EnvPrefix string

	// Environment variable handling with Viper. See [viper.SetEnvKeyReplacer]
	// for details
	EnvKeyReplacer *strings.Replacer

	viperlet *simpleviper.Viperlet

	// [*simplecommand.Command] is embedded to satisfy the
	// [simplecobra.Commander] interface
	*simplecommand.Command
}

// ensure Command satisfies the simplecobra.Commander interface
var _ simplecobra.Commander = (*Command)(nil)

// New creates a bare minimum [*Command] with a name and a short description
// set
func New(name, short string, opts ...simplecommand.CommandOption) *Command {
	return &Command{
		Command: simplecommand.New(name, short, opts...),
	}
}

// Viper allows access to the underlying [*viper.Viper] instance.
// Warning: This will return nil if called before [Command.Init] has run.
func (c *Command) Viper() *viper.Viper {
	if c.viperlet == nil {
		return nil
	}

	return c.viperlet.Viper()
}

// Init sets up environment variable and configuration file handling, then
// runs [simplecommand.Command.Init] to set the command's descriptions.
//
// If you implement your own Init method, it must call this method before
// adding any command line flags.
//
// See [simplecobra.Commander] for more information.
func (c *Command) Init(cd *simplecobra.Commandeer) error {
	// start with no options set
	opts := make([]simpleviper.Option, 0)

	// add env var handling if set
	if c.EnvPrefix != "" {
		opts = append(opts, simpleviper.WithEnvPrefix(c.EnvPrefix))
	}
	if c.EnvKeyReplacer != nil {
		opts = append(opts, simpleviper.WithEnvKeyReplacer(c.EnvKeyReplacer))
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
	if err := c.viperlet.Init(cd.CobraCommand.Flags()); err != nil {
		return err
	}

	return c.Command.Init(cd)
}

// PreRun sets any command line flags that were not provided on the command
// line from environment variables or the configuration file, then runs
// [simplecommand.Command.PreRun].
//
// If you implement your own PreRun method, it must call this method before
// using any command line flag values.
//
// See [simplecobra.Commander] for more information.
func (c *Command) PreRun(this, runner *simplecobra.Commandeer) error {
	// set any values from viper as flags once other steps are done, skipping
	// any flags set on the command line so they take precedence
	var err error
	this.CobraCommand.Flags().VisitAll(func(f *pflag.Flag) {
		if err != nil || f.Changed || !c.Viper().IsSet(f.Name) {
			return
		}

		// a list from the configuration file has no string form, so replace
		// the values of a slice flag directly
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			switch c.Viper().Get(f.Name).(type) {
			case []any, []string:
				err = sv.Replace(c.Viper().GetStringSlice(f.Name))
				return
			}
		}

		if s := c.Viper().GetString(f.Name); s != "" {
			err = this.CobraCommand.Flags().Set(f.Name, s)
		}
	})
	if err != nil {
		return err
	}

	return c.Command.PreRun(this, runner)
}
