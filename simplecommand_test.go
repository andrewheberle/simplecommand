package simplecommand_test

import (
	"slices"
	"testing"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
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
