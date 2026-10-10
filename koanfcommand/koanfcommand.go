// Package koanfcommand provides a [Command] type that, like
// [simplecommand.Command], satisfies the [simplecobra.Commander] interface but
// also allows command line flags to be set from environment variables and a
// configuration file via [koanf].
//
// Command line flags take precedence over environment variables, which take
// precedence over values from the configuration file.
//
// [koanf]: https://github.com/knadh/koanf
package koanfcommand

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// Command is the basis for creating your own [simplecobra.Commander] with
// [koanf] support and is best used by embedding it in your own struct.
//
// A [Command] should always be created using [New], as the embedded
// [*simplecommand.Command] must not be nil.
//
// [koanf]: https://github.com/knadh/koanf
type Command struct {
	// Config specifies a configuration file used. As it is read in
	// [Command.PreRun], it may be set from a command line flag.
	//
	// The file type is chosen by its extension, which may be ".yaml",
	// ".yml", ".json" or ".toml", unless ConfigParser is set.
	Config string

	// Allow missing config file when Config is set
	ConfigOptional bool

	// ConfigParser, if set, is used to parse the configuration file in place
	// of the parser chosen by the file's extension
	ConfigParser koanf.Parser

	// EnvPrefix enables setting flags from environment variables named after
	// each flag, in upper case and prefixed with EnvPrefix and an underscore.
	// For example, with an EnvPrefix of "cmd" the "example" flag is set from
	// CMD_EXAMPLE.
	EnvPrefix string

	// EnvKeyReplacer, if set, is applied to flag names to give their
	// environment variable names. For example, strings.NewReplacer("-", "_")
	// sets the "example-flag" flag from CMD_EXAMPLE_FLAG. Setting it also
	// enables environment variables when EnvPrefix is empty.
	EnvKeyReplacer *strings.Replacer

	koanf *koanf.Koanf

	// [*simplecommand.Command] is embedded to satisfy the
	// [simplecobra.Commander] interface
	*simplecommand.Command
}

// ensure Command satisfies the simplecobra.Commander interface
var _ simplecobra.Commander = (*Command)(nil)

// New creates a bare minimum [*Command] with a name and a short description
// set
func New(name, short string, opts ...simplecommand.Option) *Command {
	return &Command{
		Command: simplecommand.New(name, short, opts...),
	}
}

// Koanf allows access to the underlying [*koanf.Koanf] instance, which holds
// the contents of the configuration file. This allows reading values that
// have no command line flag, for example with [koanf.Koanf.Unmarshal].
//
// Values from environment variables are only applied to command line flags,
// so are not included.
//
// Warning: This will return nil if called before [Command.PreRun] has run.
func (c *Command) Koanf() *koanf.Koanf {
	return c.koanf
}

// PreRun reads the configuration file and sets any command line flags that
// were not provided on the command line from environment variables or the
// configuration file, then runs [simplecommand.Command.PreRun].
//
// If you implement your own PreRun method, it must call this method before
// using any command line flag values.
//
// See [simplecobra.Commander] for more information.
func (c *Command) PreRun(this, runner *simplecobra.Commandeer) error {
	c.koanf = koanf.New(".")
	if err := c.loadConfig(); err != nil {
		return err
	}

	// skip any flags set on the command line so they take precedence
	flags := this.CobraCommand.Flags()
	var err error
	flags.VisitAll(func(f *pflag.Flag) {
		if err != nil || f.Changed {
			return
		}

		if name, value, ok := c.lookupEnv(f.Name); ok {
			if err = flags.Set(f.Name, value); err != nil {
				err = fmt.Errorf("setting flag %q from environment variable %s: %w", f.Name, name, err)
			}
			return
		}

		if err = c.setFromConfig(flags, f); err != nil {
			err = fmt.Errorf("setting flag %q from configuration file: %w", f.Name, err)
		}
	})
	if err != nil {
		return err
	}

	return c.Command.PreRun(this, runner)
}

// loadConfig reads the configuration file, if set, into c.koanf
func (c *Command) loadConfig() error {
	if c.Config == "" {
		return nil
	}

	parser := c.ConfigParser
	if parser == nil {
		switch strings.ToLower(filepath.Ext(c.Config)) {
		case ".yaml", ".yml":
			parser = yaml.Parser()
		case ".json":
			parser = json.Parser()
		case ".toml":
			parser = toml.Parser()
		default:
			return fmt.Errorf("unsupported configuration file type: %s", c.Config)
		}
	}

	err := c.koanf.Load(file.Provider(c.Config), parser)
	if err != nil && c.ConfigOptional && errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}

// lookupEnv returns the name and value of the environment variable for the
// named flag. An empty value is treated as unset, as it is by viper.
func (c *Command) lookupEnv(flag string) (name, value string, ok bool) {
	if c.EnvPrefix == "" && c.EnvKeyReplacer == nil {
		return "", "", false
	}

	name = flag
	if c.EnvKeyReplacer != nil {
		name = c.EnvKeyReplacer.Replace(name)
	}
	if c.EnvPrefix != "" {
		name = c.EnvPrefix + "_" + name
	}
	name = strings.ToUpper(name)

	value, ok = os.LookupEnv(name)

	return name, value, ok && value != ""
}

// setFromConfig sets f from the configuration file, if it holds a value for it
func (c *Command) setFromConfig(flags *pflag.FlagSet, f *pflag.Flag) error {
	switch c.koanf.Get(f.Name).(type) {
	case nil:
		return nil
	case map[string]any:
		// a section of the file, which may hold values for flags such as
		// "name.key", rather than a value for this flag
		return nil
	case []any, []string:
		// a list has no string form to pass to Set, so replace the values of
		// a slice flag directly
		sv, ok := f.Value.(pflag.SliceValue)
		if !ok {
			return errors.New("a list was given for a flag that takes a single value")
		}

		return sv.Replace(c.koanf.Strings(f.Name))
	}

	if s := c.koanf.String(f.Name); s != "" {
		return flags.Set(f.Name, s)
	}

	return nil
}
