package vipercommand_test

import (
	"slices"
	"testing"

	"github.com/andrewheberle/simplecommand/vipercommand"
	"github.com/bep/simplecobra"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := &viperCommand{
				Command: vipercommand.New("example-command", "This is an example command"),
			}
			command.EnvPrefix = "cmd"
			command.Config = tt.config
			if tt.env != "" {
				t.Setenv("CMD_EXAMPLE", tt.env)
			}

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

	*vipercommand.Command
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

func TestSliceFlagNotAppended(t *testing.T) {
	command := &typedCommand{
		Command: vipercommand.New("example-command", "This is an example command"),
	}
	command.EnvPrefix = "cmd"
	t.Setenv("CMD_LIST", "a,b")

	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), []string{"--list", "z"}); err != nil {
		t.Fatal(err)
	}

	if want := []string{"z"}; !slices.Equal(command.list, want) {
		t.Errorf("got %q, want %q", command.list, want)
	}
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
			command := &typedCommand{
				Command: vipercommand.New("example-command", "This is an example command"),
			}
			command.EnvPrefix = "cmd"
			command.Config = "testconfig.yml"
			if tt.env != "" {
				t.Setenv("CMD_LIST", tt.env)
			}

			x, err := simplecobra.New(command)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := x.Execute(t.Context(), tt.args); err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(command.list, tt.want) {
				t.Errorf("got %q, want %q", command.list, tt.want)
			}
		})
	}
}

func TestInvalidEnvValue(t *testing.T) {
	command := &typedCommand{
		Command: vipercommand.New("example-command", "This is an example command"),
	}
	command.EnvPrefix = "cmd"
	t.Setenv("CMD_COUNT", "not a number")

	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), nil); err == nil {
		t.Error("expected an error for an invalid value from the environment, got nil")
	}
}

func TestConfigFromFlags(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		args    []string
		want    string
		wantErr bool
	}{
		{"config file from flag", "", []string{"--config", "testconfig.yml"}, "from config file", false},
		{"env prefix from flag", "from env var", []string{"--env-prefix", "cmd"}, "from env var", false},
		{"env beats config file from flag", "from env var", []string{"--config", "testconfig.yml", "--env-prefix", "cmd"}, "from env var", false},
		{"missing config file from flag", "", []string{"--config", "missing.yml"}, "", true},
		{"optional config file from flag", "", []string{"--config", "missing.yml", "--config-optional"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := &configFlagCommand{
				Command: vipercommand.New("example-command", "This is an example command"),
			}
			// an empty value is treated as unset
			t.Setenv("CMD_EXAMPLE", tt.env)

			x, err := simplecobra.New(command)
			if err != nil {
				t.Fatal(err)
			}
			_, err = x.Execute(t.Context(), tt.args)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("got error %v, want error %v", err, tt.wantErr)
			}

			if command.exampleFlag != tt.want {
				t.Errorf("got %q, want %q", command.exampleFlag, tt.want)
			}
		})
	}
}

func TestViper(t *testing.T) {
	command := &viperCommand{
		Command: vipercommand.New("example-command", "This is an example command"),
	}
	if v := command.Viper(); v != nil {
		t.Errorf("expected nil before PreRun, got %v", v)
	}

	command.Config = "testconfig.yml"
	x, err := simplecobra.New(command)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Execute(t.Context(), nil); err != nil {
		t.Fatal(err)
	}

	if want := "from config file"; command.Viper().GetString("example") != want {
		t.Errorf("got %q, want %q", command.Viper().GetString("example"), want)
	}
}
