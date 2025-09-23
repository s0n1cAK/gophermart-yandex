package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTooManyRequestsError(t *testing.T) {
	err := &TooManyRequestsError{RetryAfter: 5 * time.Second}
	require.Contains(t, err.Error(), "429 Too Many Requests")
	require.Contains(t, err.Error(), "5s")
}

func TestError_ErrorAndUnwrap(t *testing.T) {
	appErr := errors.New("app failed")
	srvErr := errors.New("db down")

	e := Error{AppErr: appErr, SrvErr: srvErr}
	require.Equal(t, "db down", e.Error())
	require.Equal(t, srvErr, errors.Unwrap(e))
}

func TestError_AppErrOnly(t *testing.T) {
	appErr := errors.New("only app err")
	e := Error{AppErr: appErr}
	require.Equal(t, "only app err", e.Error())
}

func TestError_UnknownError(t *testing.T) {
	e := Error{}
	require.Equal(t, "unknown error", e.Error())
}

func TestMakeError(t *testing.T) {
	appErr := errors.New("bad request")
	srvErr := errors.New("internal")

	e := MakeError(appErr, srvErr)
	require.Equal(t, appErr, e.AppErr)
	require.Equal(t, srvErr, e.SrvErr)
}

func TestWrap_WithDomainError(t *testing.T) {
	appErr := errors.New("bad request")
	dErr := MakeError(appErr, ErrServiceUnavailable)

	wrapped := Wrap("op.test", dErr)
	require.IsType(t, Error{}, wrapped)

	derr, ok := wrapped.(Error)
	require.True(t, ok)
	require.Equal(t, ErrServiceUnavailable, derr.SrvErr)
	require.Contains(t, derr.AppErr.Error(), "op.test")
}

func TestWrap_WithGenericError(t *testing.T) {
	err := errors.New("plain err")

	wrapped := Wrap("op.db", err)
	require.IsType(t, Error{}, wrapped)

	derr := wrapped.(Error)
	require.Equal(t, ErrInternal, derr.SrvErr)
	require.Contains(t, derr.AppErr.Error(), "op.db")
}

func TestGetAppErr(t *testing.T) {
	appErr := errors.New("invalid payload")
	e := MakeError(appErr, ErrInternal)

	got := GetAppErr(e)
	require.Equal(t, appErr, got)

	got = GetAppErr(errors.New("plain"))
	require.Nil(t, got)
}
