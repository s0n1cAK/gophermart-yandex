package logger

import (
	"go.uber.org/zap"
)

func New() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l, nil
}
