//go:build !darwin

package focus

import "log/slog"

type unsupportedProvider struct{}

func newProvider(_ *slog.Logger) (provider, error) {
	return nil, ErrUnsupported
}

func (unsupportedProvider) Current() (FocusSnapshot, error) {
	return FocusSnapshot{}, ErrUnsupported
}
