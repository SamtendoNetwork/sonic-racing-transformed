package main

import (
	"testing"

	"github.com/nex-server/sonic-transformed/globals"
)

func TestAccountsConfigurationRequired(t *testing.T) {
	old := globals.Settings
	t.Cleanup(func() { globals.Settings = old })
	valid := globals.Config{
		AuthenticationServerPort: 60000, SecureServerPort: 60001,
		SecureServerHost: "127.0.0.1", AccessKey: globals.DefaultAccessKey,
		AccountGRPCHost: "accounts", AccountGRPCPort: 50051, AccountGRPCAPIKey: "test-key",
	}
	globals.Settings = valid
	if err := validateConfig(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*globals.Config)
	}{
		{"missing host", func(c *globals.Config) { c.AccountGRPCHost = "" }},
		{"missing port", func(c *globals.Config) { c.AccountGRPCPort = 0 }},
		{"invalid port", func(c *globals.Config) { c.AccountGRPCPort = 65536 }},
		{"missing key", func(c *globals.Config) { c.AccountGRPCAPIKey = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			globals.Settings = valid
			tc.change(&globals.Settings)
			if err := validateConfig(); err == nil {
				t.Fatal("accepted incomplete accounts configuration")
			}
		})
	}
}
