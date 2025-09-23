package gophermart

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateENV(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "http://test.test:8080")
	t.Setenv("DATABASE_URI", "http://test.database")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual.local:9000")
	t.Setenv("ENVIRONMENT", "dev")

	cfg, err := New()
	require.NoError(t, err)

	require.Equal(t, "http://test.test:8080", cfg.Address.String())
	require.Equal(t, "http://test.database", cfg.DatabaseURI.String())
	require.Equal(t, "http://accrual.local:9000", cfg.AccuralAddress.String())
	require.Equal(t, "dev", cfg.Environment)
}

func TestValidateFlags(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	_ = os.Unsetenv("RUN_ADDRESS")
	_ = os.Unsetenv("DATABASE_URI")
	_ = os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")

	os.Args = []string{"cmd",
		"-a", "http://testflag.test",
		"-d", "http://testflag.database",
		"-r", "testflagfile",
	}

	_, err := os.Create("testflagfile")
	require.NoError(t, err)
	defer os.Remove("testflagfile")

	cfg, err := New()
	require.NoError(t, err)

	require.Equal(t, cfg.Address.String(), "http://testflag.test")
	require.Equal(t, cfg.DatabaseURI.String(), "http://testflag.database")
	require.Equal(t, cfg.Accrual, "testflagfile")
}

func TestMetamorphosis(t *testing.T) {
	tests := []struct {
		name    string
		cfg     initConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: initConfig{
				Address:        "http://localhost:8080",
				DatabaseURI:    "http://test.db",
				Accrual:        "/bin/accrual",
				Environment:    "dev",
				AccuralAddress: "http://accrual.local:9000",
			},
			wantErr: false,
		},
		{
			name: "missing scheme in address",
			cfg: initConfig{
				Address:        "localhost:8080",
				DatabaseURI:    "http://test.db",
				Accrual:        "/bin/accrual",
				Environment:    "dev",
				AccuralAddress: "http://accrual.local:9000",
			},
			wantErr: false,
		},
		{
			name: "invalid address host",
			cfg: initConfig{
				Address:        "http://:8080",
				DatabaseURI:    "http://test.db",
				Accrual:        "/bin/accrual",
				Environment:    "dev",
				AccuralAddress: "http://accrual.local:9000",
			},
			wantErr: true,
		},
		{
			name: "invalid port in address",
			cfg: initConfig{
				Address:        "http://localhost:99999",
				DatabaseURI:    "http://test.db",
				Accrual:        "/bin/accrual",
				Environment:    "dev",
				AccuralAddress: "http://accrual.local:9000",
			},
			wantErr: true,
		},
		{
			name: "invalid database URI",
			cfg: initConfig{
				Address:        "http://localhost:8080",
				DatabaseURI:    "://broken.db",
				Accrual:        "/bin/accrual",
				Environment:    "dev",
				AccuralAddress: "http://accrual.local:9000",
			},
			wantErr: true,
		},
		{
			name: "invalid accrual address",
			cfg: initConfig{
				Address:        "http://localhost:8080",
				DatabaseURI:    "http://test.db",
				Accrual:        "/bin/accrual",
				Environment:    "dev",
				AccuralAddress: "://broken",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			cfg: initConfig{
				Address:        "http://localhost:8080",
				DatabaseURI:    "http://test.db",
				Accrual:        "/bin/accrual",
				Environment:    "stage",
				AccuralAddress: "http://accrual.local:9000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := metamorphosis(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg.Address)
				require.NotNil(t, cfg.DatabaseURI)
				require.NotNil(t, cfg.AccuralAddress)
				require.NotEmpty(t, cfg.Accrual)
				require.Contains(t, []string{"prod", "dev", "test"}, cfg.Environment)
			}
		})
	}
}
