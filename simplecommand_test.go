package simplecommand_test

import (
	"slices"
	"testing"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
	"github.com/spf13/cobra"
)

// names returns the names of the provided commands
func names(commands []simplecobra.Commander) []string {
	n := make([]string, 0, len(commands))
	for _, c := range commands {
		n = append(n, c.Name())
	}

	return n
}

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		opts           []simplecommand.Option
		wantLong       string
		wantDeprecated string
		wantCommands   []string
	}{
		{"no options", nil, "", "", []string{}},
		{"with long", []simplecommand.Option{simplecommand.WithLong("long description")}, "long description", "", []string{}},
		{"with deprecated", []simplecommand.Option{simplecommand.WithDeprecated("no longer used")}, "", "no longer used", []string{}},
		{
			"with subcommands",
			[]simplecommand.Option{simplecommand.WithSubCommands(
				simplecommand.New("sub-one", "first"),
				simplecommand.New("sub-two", "second"),
			)},
			"", "", []string{"sub-one", "sub-two"},
		},
		{
			"with all options",
			[]simplecommand.Option{
				simplecommand.WithLong("long description"),
				simplecommand.WithDeprecated("no longer used"),
				simplecommand.WithSubCommands(simplecommand.New("sub-one", "first")),
			},
			"long description", "no longer used", []string{"sub-one"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := simplecommand.New("example-command", "short description", tt.opts...)

			if got := c.Name(); got != "example-command" {
				t.Errorf("Name() = %q, want %q", got, "example-command")
			}
			if c.Short != "short description" {
				t.Errorf("Short = %q, want %q", c.Short, "short description")
			}
			if c.Long != tt.wantLong {
				t.Errorf("Long = %q, want %q", c.Long, tt.wantLong)
			}
			if c.Deprecated != tt.wantDeprecated {
				t.Errorf("Deprecated = %q, want %q", c.Deprecated, tt.wantDeprecated)
			}
			if got := names(c.Commands()); !slices.Equal(got, tt.wantCommands) {
				t.Errorf("Commands() = %q, want %q", got, tt.wantCommands)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name string
		opts []simplecommand.Option
	}{
		{"defaults", nil},
		{"with long", []simplecommand.Option{simplecommand.WithLong("long description")}},
		{"with deprecated", []simplecommand.Option{simplecommand.WithDeprecated("no longer used")}},
		{"with aliases", []simplecommand.Option{simplecommand.WithAliases("sub", "sc")}},
		{"with example", []simplecommand.Option{simplecommand.WithExample("root-command sub-command")}},
		{"with args", []simplecommand.Option{simplecommand.WithArgs(cobra.NoArgs)}},
		{"with hidden", []simplecommand.Option{simplecommand.WithHidden()}},
		{"with version", []simplecommand.Option{simplecommand.WithVersion("1.2.3")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := simplecommand.New("sub-command", "short description", tt.opts...)
			root := simplecommand.New("root-command", "root command", simplecommand.WithSubCommands(sub))

			x, err := simplecobra.New(root)
			if err != nil {
				t.Fatal(err)
			}

			cd, err := x.Execute(t.Context(), []string{"sub-command"})
			if err != nil {
				t.Fatal(err)
			}

			// the default Init should copy the descriptions to the cobra command
			cmd := cd.CobraCommand
			if cmd.Name() != sub.Name() {
				t.Fatalf("ran %q, want %q", cmd.Name(), sub.Name())
			}
			if cmd.Short != sub.Short {
				t.Errorf("Short = %q, want %q", cmd.Short, sub.Short)
			}
			if cmd.Long != sub.Long {
				t.Errorf("Long = %q, want %q", cmd.Long, sub.Long)
			}
			if cmd.Deprecated != sub.Deprecated {
				t.Errorf("Deprecated = %q, want %q", cmd.Deprecated, sub.Deprecated)
			}
			if !slices.Equal(cmd.Aliases, sub.Aliases) {
				t.Errorf("Aliases = %q, want %q", cmd.Aliases, sub.Aliases)
			}
			if cmd.Example != sub.Example {
				t.Errorf("Example = %q, want %q", cmd.Example, sub.Example)
			}
			if (cmd.Args == nil) != (sub.Args == nil) {
				t.Errorf("Args set = %v, want %v", cmd.Args != nil, sub.Args != nil)
			}
			if cmd.Hidden != sub.Hidden {
				t.Errorf("Hidden = %v, want %v", cmd.Hidden, sub.Hidden)
			}
			if cmd.Version != sub.Version {
				t.Errorf("Version = %q, want %q", cmd.Version, sub.Version)
			}
		})
	}
}

func TestWithSubCommandsAppends(t *testing.T) {
	c := simplecommand.New("example-command", "short description",
		simplecommand.WithSubCommands(simplecommand.New("sub-one", "first")),
		simplecommand.WithSubCommands(
			simplecommand.New("sub-two", "second"),
			simplecommand.New("sub-three", "third"),
		),
	)

	want := []string{"sub-one", "sub-two", "sub-three"}
	if got := names(c.Commands()); !slices.Equal(got, want) {
		t.Errorf("Commands() = %q, want %q", got, want)
	}
}

func TestWithSubCommandsCopies(t *testing.T) {
	subcommands := []simplecobra.Commander{simplecommand.New("sub-one", "first")}
	c := simplecommand.New("example-command", "short description", simplecommand.WithSubCommands(subcommands...))

	// changing the caller's slice afterwards must not change the command
	subcommands[0] = simplecommand.New("changed", "changed")

	want := []string{"sub-one"}
	if got := names(c.Commands()); !slices.Equal(got, want) {
		t.Errorf("Commands() = %q, want %q", got, want)
	}
}

func TestWithAliasesAppends(t *testing.T) {
	c := simplecommand.New("example-command", "short description",
		simplecommand.WithAliases("one"),
		simplecommand.WithAliases("two", "three"),
	)

	if want := []string{"one", "two", "three"}; !slices.Equal(c.Aliases, want) {
		t.Errorf("Aliases = %q, want %q", c.Aliases, want)
	}
}

func TestWithAliasesRuns(t *testing.T) {
	sub := simplecommand.New("sub-command", "short description", simplecommand.WithAliases("sc"))
	root := simplecommand.New("root-command", "root command", simplecommand.WithSubCommands(sub))

	x, err := simplecobra.New(root)
	if err != nil {
		t.Fatal(err)
	}

	cd, err := x.Execute(t.Context(), []string{"sc"})
	if err != nil {
		t.Fatal(err)
	}

	if cd.CobraCommand.Name() != sub.Name() {
		t.Errorf("ran %q, want %q", cd.CobraCommand.Name(), sub.Name())
	}
}

func TestWithArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", nil, true},
		{"one arg", []string{"one"}, false},
		{"two args", []string{"one", "two"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := simplecommand.New("example-command", "short description", simplecommand.WithArgs(cobra.ExactArgs(1)))

			x, err := simplecobra.New(c)
			if err != nil {
				t.Fatal(err)
			}

			_, err = x.Execute(t.Context(), tt.args)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("got error %v, want error %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeprecatedAliases(t *testing.T) {
	// the deprecated names must behave the same as their replacements
	opts := []simplecommand.CommandOption{
		simplecommand.Long("long description"),
		simplecommand.Deprecated("no longer used"),
	}

	got := simplecommand.New("example-command", "short description", opts...)
	want := simplecommand.New("example-command", "short description",
		simplecommand.WithLong("long description"),
		simplecommand.WithDeprecated("no longer used"),
	)

	if got.Long != want.Long {
		t.Errorf("Long = %q, want %q", got.Long, want.Long)
	}
	if got.Deprecated != want.Deprecated {
		t.Errorf("Deprecated = %q, want %q", got.Deprecated, want.Deprecated)
	}
}
