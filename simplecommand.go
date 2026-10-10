// Package simplecommand reduces the amount of boilerplate code required to
// use [simplecobra] as it provides a [Command] type that satisfies the
// [simplecobra.Commander] that you can embed within your own custom type and
// implement your own [Command.Init], [Command.PreRun] and [Command.Run]
// methods as required.
package simplecommand

import (
	"context"

	"github.com/bep/simplecobra"
	"github.com/spf13/cobra"
)

// Command is the basis for creating your own [simplecobra.Commander] quickly.
// A [Command] satisfies the [simplecobra.Commander] interface and is best
// used by embedding it in your own struct.
//
// When embedding, always set the embedded *Command (for example using [New]),
// as a nil *Command will cause a panic when its methods are called.
type Command struct {
	// CommandName is used as the command's name for any help pages
	CommandName string

	// Short and Long are the command's short and long descriptions for help
	// pages, and Deprecated marks the command as deprecated with the given
	// reason. These are set on the command by the default Init method, so
	// when implementing your own Init method call [Command.Init] from it, or
	// set these yourself.
	Short      string
	Long       string
	Deprecated string

	// Aliases, Example, Args, Hidden and Version are set on the command when
	// using the default Init method. See [cobra.Command] for details of each.
	Aliases []string
	Example string
	Args    cobra.PositionalArgs
	Hidden  bool
	Version string

	// SubCommands holds the list of sub-commands for this command
	SubCommands []simplecobra.Commander
}

// ensure Command satisfies the simplecobra.Commander interface
var _ simplecobra.Commander = (*Command)(nil)

// New creates a bare minimum [Command] with a name and a short description
// set
func New(name, short string, opts ...Option) *Command {
	c := &Command{
		CommandName: name,
		Short:       short,
	}

	// set options
	for _, o := range opts {
		o(c)
	}

	return c
}

// Name returns the name of the command.
func (c *Command) Name() string {
	return c.CommandName
}

// Commands returns the sub-commands of the command.
func (c *Command) Commands() []simplecobra.Commander {
	return c.SubCommands
}

// Init applies the command's descriptions and other settings, such as those
// set by [WithArgs] or [WithAliases], to the underlying [cobra.Command].
//
// The default adds no command line flags, so is suitable as is for a command
// that has none. To add flags, implement your own Init method that calls this
// one and then adds them, as shown in the examples for [New].
//
// See [simplecobra.Commander] for more information.
func (c *Command) Init(cd *simplecobra.Commandeer) error {
	cmd := cd.CobraCommand
	cmd.Short = c.Short
	cmd.Long = c.Long
	cmd.Deprecated = c.Deprecated
	cmd.Aliases = c.Aliases
	cmd.Example = c.Example
	cmd.Args = c.Args
	cmd.Hidden = c.Hidden
	cmd.Version = c.Version

	return nil
}

// PreRun is where command line flags have been parsed, so is a place for any
// initialisation would go for the command.
// The default is only suitable for implementing a command that has no reliance
// on internal state such as command line flags.
//
// See [simplecobra.Commander] for more information.
func (c *Command) PreRun(this, runner *simplecobra.Commandeer) error {
	return nil
}

// Run is where the command actually does its work.
// The default does no actual work, so is likely not suitable for any use case
// except for possibly a deprecated command.
//
// See [simplecobra.Commander] for more information.
func (c *Command) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	return nil
}

// An Option is passed to [New] to change the defaults of the [Command]
type Option func(*Command)

// WithLong sets the long description of the command when the default
// [Command.Init] is used.
func WithLong(description string) Option {
	return func(c *Command) {
		c.Long = description
	}
}

// WithDeprecated sets command as deprecated when the default [Command.Init] is used.
func WithDeprecated(reason string) Option {
	return func(c *Command) {
		c.Deprecated = reason
	}
}

// WithSubCommands adds sub-commands to the command during [New]. It may be
// passed more than once, with each call adding to any existing sub-commands.
func WithSubCommands(subcommands ...simplecobra.Commander) Option {
	return func(c *Command) {
		c.SubCommands = append(c.SubCommands, subcommands...)
	}
}

// WithAliases adds aliases that may be used in place of the command's name
// when the default [Command.Init] is used. It may be passed more than once,
// with each call adding to any existing aliases.
func WithAliases(aliases ...string) Option {
	return func(c *Command) {
		c.Aliases = append(c.Aliases, aliases...)
	}
}

// WithExample sets examples of how to use the command, which are shown in its
// help when the default [Command.Init] is used.
func WithExample(example string) Option {
	return func(c *Command) {
		c.Example = example
	}
}

// WithArgs sets the validation of the command's positional arguments, such
// as [cobra.ExactArgs] or [cobra.NoArgs], when the default [Command.Init] is
// used. Invalid arguments return an error before [Command.PreRun] runs.
func WithArgs(args cobra.PositionalArgs) Option {
	return func(c *Command) {
		c.Args = args
	}
}

// WithHidden hides the command from the list of available commands in help
// output when the default [Command.Init] is used. The command can still be
// run.
func WithHidden() Option {
	return func(c *Command) {
		c.Hidden = true
	}
}

// WithVersion sets the command's version when the default [Command.Init] is
// used. This adds a --version flag (and -v, if not already used) that
// prints it, so is intended for the root command.
func WithVersion(version string) Option {
	return func(c *Command) {
		c.Version = version
	}
}

// CommandOption is the previous name of [Option].
//
// Deprecated: Use [Option].
type CommandOption = Option

// Long sets the long description of the command when the default
// [Command.Init] is used.
//
// Deprecated: Use [WithLong].
func Long(description string) Option {
	return WithLong(description)
}

// Deprecated sets command as deprecated when the default [Command.Init] is
// used.
//
// Deprecated: Use [WithDeprecated].
func Deprecated(reason string) Option {
	return WithDeprecated(reason)
}
