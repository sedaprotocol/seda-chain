package utils

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cast"

	tmcfg "github.com/cometbft/cometbft/config"

	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
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

# enable-seda-signer enables the SEDA signer.
enable-seda-signer = {{ .SEDAConfig.EnableSEDASigner }}

# seda-key-file is the path to the SEDA key file from the node's home directory.
seda-key-file = "{{ .SEDAConfig.SEDAKeyFile }}"

# allow-unencrypted-seda-keys enables unencrypted use of the SEDA key file.
allow-unencrypted-seda-keys = {{ .SEDAConfig.AllowUnencryptedSEDAKeys }}
`
)

const (
	FlagEnableSEDASigner         = "seda.enable-seda-signer"
	FlagSEDAKeyFile              = "seda.seda-key-file"
	FlagAllowUnencryptedSEDAKeys = "seda.allow-unencrypted-seda-keys"
)

var defaultSEDAKeyFile = filepath.Join(tmcfg.DefaultConfigDir, "seda_keys.json")

// AppConfig defines the application configurations. It extends the default
// Cosmos SDK server config with custom SEDA configurations.
type AppConfig struct {
	serverconfig.Config
	SEDAConfig SEDAConfig `mapstructure:"seda_config"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Config:     *serverconfig.DefaultConfig(),
		SEDAConfig: DefaultSEDAConfig(),
	}
}

type SEDAConfig struct {
	EnableSEDASigner         bool   `mapstructure:"enable-seda-signer"`
	SEDAKeyFile              string `mapstructure:"seda-key-file"`
	AllowUnencryptedSEDAKeys bool   `mapstructure:"allow-unencrypted-seda-keys"`
}

func DefaultSEDAConfig() SEDAConfig {
	return SEDAConfig{
		EnableSEDASigner:         true,
		SEDAKeyFile:              defaultSEDAKeyFile,
		AllowUnencryptedSEDAKeys: false,
	}
}

// ReadSEDAConfigFromAppOpts parses an AppOptions and returns a SEDAConfig. It
// returns an error if there is a missing configuration.
func ReadSEDAConfigFromAppOpts(appOpts servertypes.AppOptions) (*SEDAConfig, error) {
	config := new(SEDAConfig)

	v := appOpts.Get(FlagEnableSEDASigner)
	if v == nil {
		return nil, fmt.Errorf("%s is not configured in app.toml", FlagEnableSEDASigner)
	}
	config.EnableSEDASigner = cast.ToBool(v)

	v = appOpts.Get(FlagSEDAKeyFile)
	if v == nil {
		return nil, fmt.Errorf("%s is not configured in app.toml", FlagSEDAKeyFile)
	}
	config.SEDAKeyFile = cast.ToString(v)

	v = appOpts.Get(FlagAllowUnencryptedSEDAKeys)
	if v == nil {
		return nil, fmt.Errorf("%s is not configured in app.toml", FlagAllowUnencryptedSEDAKeys)
	}
	config.AllowUnencryptedSEDAKeys = cast.ToBool(v)

	return config, nil
}

// configKeyNodeStart is added to the configuration at the PreRun of the start
// command to indicate node start.
const configKeyNodeStart = "nodeStart"

func SetNodeStart(serverCtx *server.Context) {
	serverCtx.Viper.Set(configKeyNodeStart, true)
}

func IsNodeStart(appOpts servertypes.AppOptions) bool {
	return appOpts.Get(configKeyNodeStart) != nil
}
