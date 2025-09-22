package gophermart

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateENV(t *testing.T) {
	_ = os.Unsetenv("RUN_ADDRESS")
	_ = os.Unsetenv("DATABASE_URI")
	_ = os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")

	os.Setenv("RUN_ADDRESS", "http://test.test")
	os.Setenv("DATABASE_URI", "http://test.database")
	os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "testfile")

	_, err := os.Create("testfile")
	require.NoError(t, err)
	defer os.Remove("testfile")

	cfg, err := New()
	require.NoError(t, err)

	require.Equal(t, cfg.Address, "http://test.test")
	require.Equal(t, cfg.DatabaseURI, "http://test.database")
	require.Equal(t, cfg.Accrual, "testfile")
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

	require.Equal(t, cfg.Address, "http://testflag.test")
	require.Equal(t, cfg.DatabaseURI, "http://testflag.database")
	require.Equal(t, cfg.Accrual, "testflagfile")
}
