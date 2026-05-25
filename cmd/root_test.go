package cmd

import (
	"testing"

	"github.com/chef/go-libs/credentials"
	"github.com/stretchr/testify/require"
)

func TestHasMinimumParams(t *testing.T) {
	orig := infraFlags
	t.Cleanup(func() {
		infraFlags = orig
	})

	infraFlags.chefServerURL = "https://example.chef"
	infraFlags.clientName = "tester"
	infraFlags.clientKey = "/tmp/key.pem"
	require.True(t, hasMinimumParams())

	infraFlags.clientKey = ""
	require.False(t, hasMinimumParams())
}

func TestOverrideCredentials(t *testing.T) {
	orig := infraFlags
	t.Cleanup(func() {
		infraFlags = orig
	})

	base := &credentials.Credentials{
		CredsDetail: credentials.CredsDetail{
			ClientName:    "orig-name",
			ClientKey:     "orig-key",
			ChefServerUrl: "https://orig.example",
		},
	}

	infraFlags.clientName = "new-name"
	infraFlags.clientKey = "new-key"
	infraFlags.chefServerURL = "https://new.example"

	overrideCredentials()(base)
	require.Equal(t, "new-name", base.ClientName)
	require.Equal(t, "new-key", base.ClientKey)
	require.Equal(t, "https://new.example", base.ChefServerUrl)

	base = &credentials.Credentials{
		CredsDetail: credentials.CredsDetail{
			ClientName:    "keep-name",
			ClientKey:     "keep-key",
			ChefServerUrl: "https://keep.example",
		},
	}
	infraFlags.clientName = ""
	infraFlags.clientKey = ""
	infraFlags.chefServerURL = ""

	overrideCredentials()(base)
	require.Equal(t, "keep-name", base.ClientName)
	require.Equal(t, "keep-key", base.ClientKey)
	require.Equal(t, "https://keep.example", base.ChefServerUrl)
}
