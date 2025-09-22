package gophermart

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	minPort = 0
	maxPort = 65535
)

func metamorphosis(cfg initConfig) (Config, error) {
	op := "config.metamorphosis"

	if !strings.Contains(cfg.Address, "://") {
		cfg.Address = "http://" + cfg.Address
	}
	address, err := url.ParseRequestURI(cfg.Address)
	if err != nil {
		return Config{}, fmt.Errorf("%s: invalid server address: %w", op, err)
	}

	if address.Host == "" {
		return Config{}, fmt.Errorf("%s: invalid server address: missing scheme or host", op)
	}

	if !strings.Contains(cfg.Address, "://") {
		cfg.AccuralAddress = "http://" + cfg.AccuralAddress
	}
	accurualAddress, err := url.ParseRequestURI(cfg.AccuralAddress)
	if err != nil {
		return Config{}, fmt.Errorf("%s: invalid accurual address: %w", op, err)
	}
	if accurualAddress.Scheme == "" {
		accurualAddress.Scheme = "http"
	}
	if accurualAddress.Host == "" {
		return Config{}, fmt.Errorf("%s: invalid accurual address: missing scheme or host", op)
	}

	port, err := strconv.Atoi(address.Port())
	if err != nil {
		return Config{}, fmt.Errorf("%s: invalid server address: bad port %q", op, address.Port())
	}
	if port < minPort || port > maxPort {
		return Config{}, fmt.Errorf("%s: invalid server address: port out of range (%d)", op, port)
	}

	database, err := url.ParseRequestURI(cfg.DatabaseURI)
	if err != nil {
		return Config{}, fmt.Errorf("%s: invalid server database: %w", op, err)
	}
	if database.Scheme == "" || database.Host == "" {
		return Config{}, fmt.Errorf("%s: invalid server database: missing scheme or host", op)
	}

	if cfg.Environment != "prod" && cfg.Environment != "dev" && cfg.Environment != "test" {
		return Config{}, fmt.Errorf("%s: unknown environment %q", op, cfg.Environment)
	}

	return Config{Address: address, DatabaseURI: database, Accrual: cfg.Accrual, Environment: cfg.Environment, AccuralAddress: accurualAddress}, nil
}
