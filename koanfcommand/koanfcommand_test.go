package koanfcommand_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/andrewheberle/simplecommand/koanfcommand"
	"github.com/bep/simplecobra"
	"github.com/knadh/koanf/parsers/json"
)

func TestFlagPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		config string
		args   []string
		want   string
	}{
		{"flag beats env", "from env var", "", []string{"--example", "from command line"}, "from command line"},
		{"flag beats config", "", "testconfig.yml", []string{"--example", "from command line"}, "from command line"},
		{"env beats config", "from env var", "testconfig.yml", nil, "from env var"},
		{"config used when nothing else set", "", "testconfig.yml", nil, "from config file"},
		{"default used when nothing set", "", "", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := &koanfCommand{
				Command: koanfcommand.New("example-command", "This is an example command"),
			}
			command.EnvPrefix = "cmd"
			command.Config = tt.config
			// an empty value is treated as unset
			t.Setenv("CMD_EXAMPLE", tt.env)

			x, err := simplecobra.New(command)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := x.Execute(t.Context(), tt.args); err != nil {
				t.Fatal(err)
			}

			if command.exampleFlag != tt.want {
				t.Errorf("got %q, want %q", command.exampleFlag, tt.want)
			}
		})
	}
}

type typedCommand struct {
	list  []string
	count int

	*koanfcommand.Command
}

func (c *typedCommand) Init(cd *simplecobra.Commandeer) error {
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	cmd := cd.CobraCommand
	cmd.Flags().StringSliceVar(&c.list, "list", nil, "Example slice flag")
	cmd.Flags().IntVar(&c.count, "count", 0, "Example int flag")

	return nil
}

func runTyped(t *testing.T, config string, args []string) (*typedCommand, error) {
	t.Helper()

	command := &typedCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	command.EnvPrefix = "cmd"
	command.Config = config

	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.Execute(t.Context(), args)

	return command, err
}

func TestSliceFlagPrecedence(t *testing.T) {
	tests := []struct {
		name string
		env  string
		args []string
		want []string
	}{
		{"flag beats config", "", []string{"--list", "z"}, []string{"z"}},
		{"env beats config", "x,y", nil, []string{"x", "y"}},
		{"config list used when nothing else set", "", nil, []string{"from", "config"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CMD_LIST", tt.env)

			command, err := runTyped(t, "testconfig.yml", tt.args)
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(command.list, tt.want) {
				t.Errorf("got %q, want %q", command.list, tt.want)
			}
		})
	}
}

func TestSliceFlagNotAppended(t *testing.T) {
	t.Setenv("CMD_LIST", "a,b")

	command, err := runTyped(t, "", []string{"--list", "z"})
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{"z"}; !slices.Equal(command.list, want) {
		t.Errorf("got %q, want %q", command.list, want)
	}
}

func TestConfigFormats(t *testing.T) {
	tests := []struct {
		config    string
		wantList  []string
		wantCount int
	}{
		{"testconfig.yml", []string{"from", "config"}, 0},
		{"testconfig.json", nil, 2},
		{"testconfig.toml", []string{"from", "toml"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.config, func(t *testing.T) {
			command, err := runTyped(t, tt.config, nil)
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(command.list, tt.wantList) || command.count != tt.wantCount {
				t.Errorf("got list %q and count %d, want list %q and count %d",
					command.list, command.count, tt.wantList, tt.wantCount)
			}
		})
	}
}

func TestConfigErrors(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		optional bool
		wantErr  bool
	}{
		{"missing config", "missing.yml", false, true},
		{"missing optional config", "missing.yml", true, false},
		{"unsupported extension", "testconfig.ini", false, true},
		{"unsupported extension when optional", "testconfig.ini", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := &typedCommand{
				Command: koanfcommand.New("example-command", "This is an example command"),
			}
			command.Config = tt.config
			command.ConfigOptional = tt.optional

			x, err := simplecobra.New(command)
			if err != nil {
				t.Fatal(err)
			}
			_, err = x.Execute(t.Context(), nil)

			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("got error %v, want error %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigParser(t *testing.T) {
	command := &typedCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	// a YAML file is not valid JSON, so this fails if ConfigParser is used
	command.Config = "testconfig.yml"
	command.ConfigParser = json.Parser()

	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), nil); err == nil {
		t.Error("expected an error parsing a YAML file with the JSON parser, got nil")
	}
}

func TestInvalidEnvValue(t *testing.T) {
	t.Setenv("CMD_COUNT", "not a number")

	if _, err := runTyped(t, "", nil); err == nil {
		t.Error("expected an error for an invalid value from the environment, got nil")
	}
}

// stringListCommand has a string flag named "list", which testconfig.yml
// sets to a list
type stringListCommand struct {
	list string

	*koanfcommand.Command
}

func (c *stringListCommand) Init(cd *simplecobra.Commandeer) error {
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	cd.CobraCommand.Flags().StringVar(&c.list, "list", "", "Example string flag")

	return nil
}

func TestListForSingleValueFlag(t *testing.T) {
	command := &stringListCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	command.Config = "testconfig.yml"

	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), nil); err == nil {
		t.Error("expected an error for a list given for a single value flag, got nil")
	}
}

func TestEnvKeyReplacerWithoutPrefix(t *testing.T) {
	command := &koanfCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	command.EnvKeyReplacer = strings.NewReplacer(".", "_")
	t.Setenv("EXAMPLE_FLAG", "from env var")

	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), nil); err != nil {
		t.Fatal(err)
	}

	if want := "from env var"; command.exampleLongFlag != want {
		t.Errorf("got %q, want %q", command.exampleLongFlag, want)
	}
}

func TestNestedConfig(t *testing.T) {
	command := &koanfCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	// a section in the file sets the "example.flag" flag and is not used for
	// the "example" flag
	command.Config = "testnested.yml"

	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), nil); err != nil {
		t.Fatal(err)
	}

	if command.exampleFlag != "" || command.exampleLongFlag != "from section" {
		t.Errorf("got %q and %q, want %q and %q", command.exampleFlag, command.exampleLongFlag, "", "from section")
	}
}

func TestKoanf(t *testing.T) {
	command := &koanfCommand{
		Command: koanfcommand.New("example-command", "This is an example command"),
	}
	if k := command.Koanf(); k != nil {
		t.Errorf("expected nil before PreRun, got %v", k)
	}

	command.Config = "testconfig.yml"
	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), nil); err != nil {
		t.Fatal(err)
	}

	if want := []string{"from", "config"}; !slices.Equal(command.Koanf().Strings("list"), want) {
		t.Errorf("got %q, want %q", command.Koanf().Strings("list"), want)
	}
}
