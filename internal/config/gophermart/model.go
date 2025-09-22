package gophermart

import "net/url"

type initConfig struct {
	Address        string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	Accrual        string `env:"ACCURUAL_ADDRESS"`
	Environment    string `env:"ENVIRONMENT"`
	AccuralAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type Config struct {
	Address        *url.URL
	DatabaseURI    *url.URL
	Accrual        string
	Environment    string
	AccuralAddress *url.URL
}
