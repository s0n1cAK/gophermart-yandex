package job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThrottler(t *testing.T) {
	th := NewThrottler()
	require.False(t, th.IsPaused())

	th.Pause(50 * time.Millisecond)
	require.True(t, th.IsPaused())

	time.Sleep(60 * time.Millisecond)
	require.False(t, th.IsPaused())
}
