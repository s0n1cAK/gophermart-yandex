package lib

import "fmt"

func StandardError(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
