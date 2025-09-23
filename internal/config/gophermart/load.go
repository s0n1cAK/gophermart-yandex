package gophermart

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

func NewDefaultConfig() initConfig {
	return initConfig{
		Address:        "http://localhost:8080",
		DatabaseURI:    "postgres://test_user:test_password@localhost:5432/gophermart?sslmode=disable",
		Accrual:        "cmd/accrual/accrual_linux_amd64",
		Environment:    "dev",
		AccuralAddress: "http://localhost:8085",
	}
}

func makeConfig(fs *flag.FlagSet, args []string) (Config, error) {
	op := "config.make"

	defaultCfg := NewDefaultConfig()

	if err := env.Parse(&defaultCfg); err != nil {
		return Config{}, fmt.Errorf("%s: %w", op, err)
	}

	fs.StringVar(&defaultCfg.Address, "a", defaultCfg.Address, "Server address")
	fs.StringVar(&defaultCfg.DatabaseURI, "d", defaultCfg.DatabaseURI, "Database URI")
	fs.StringVar(&defaultCfg.Accrual, "r", defaultCfg.Accrual, "Path to accural app")
	fs.StringVar(&defaultCfg.Environment, "e", defaultCfg.Environment, "Environment")
	fs.StringVar(&defaultCfg.AccuralAddress, "z", defaultCfg.AccuralAddress, "Accurual server address")

	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("%s: %w", op, err)
	}

	cfg, err := metamorphosis(defaultCfg)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Printf("%+v", cfg)

	return cfg, nil
}

func New() (Config, error) {
	return makeConfig(flag.CommandLine, os.Args[1:])
}
