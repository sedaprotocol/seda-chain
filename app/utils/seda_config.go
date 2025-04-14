package utils

import (
	"path/filepath"

	"github.com/spf13/cast"

	tmcfg "github.com/cometbft/cometbft/config"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

const (
	// DefaultSEDATemplate should be added to the app.toml file for SEDA-specific
	// configurations.
	DefaultSEDATemplate = `

###############################################################################
###                                   SEDA                                  ###
###############################################################################

[seda]

# seda_key_file is the path to the SEDA key file from the node's home directory.
seda_key_file = "{{ .SEDAConfig.SEDAKey }}"
`
)

var defaultSEDAKeyPath = filepath.Join(tmcfg.DefaultConfigDir, "seda_keys.json")

type SEDAConfig struct {
	SEDAKey string `mapstructure:"seda_key_file"`
}

func DefaultSEDAConfig() SEDAConfig {
	return SEDAConfig{
		SEDAKey: defaultSEDAKeyPath,
	}
}

// GetSEDAConfig parses an AppOptions and returns a SEDAConfig. If the "seda"
// key is not present in the AppOptions, it returns a default value.
func GetSEDAConfig(appOpts servertypes.AppOptions) SEDAConfig {
	v := appOpts.Get("seda")
	if v == nil {
		return DefaultSEDAConfig()
	}
	configMap := cast.ToStringMapString(v)
	return SEDAConfig{
		SEDAKey: configMap["seda_key_file"],
	}
}
