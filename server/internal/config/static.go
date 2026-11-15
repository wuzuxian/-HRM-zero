package config

import "github.com/jzero-io/jzero/core/configcenter"

// staticConfigCenter is a ConfigCenter holding a fixed config value. Used by
// logic tests to inject config without a file-backed subscriber.
type staticConfigCenter struct {
	value Config
}

func (s staticConfigCenter) GetConfig() (Config, error) { return s.value, nil }

func (s staticConfigCenter) MustGetConfig() Config { return s.value }

func (s staticConfigCenter) AddListener(func()) {}

// NewStaticConfigCenter returns a ConfigCenter that always returns the given
// config (test seam; mirrors jzero configcenter.ConfigCenter).
func NewStaticConfigCenter(value Config) configcenter.ConfigCenter[Config] {
	return staticConfigCenter{value: value}
}
