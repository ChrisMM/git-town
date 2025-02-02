package configfile

import (
	"fmt"
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/git-town/git-town/v11/src/config/configdomain"
	"github.com/git-town/git-town/v11/src/git/gitdomain"
	"github.com/git-town/git-town/v11/src/messages"
)

// Decode converts the given config file TOML source into Go data.
func Decode(text string) (*Data, error) {
	var result Data
	_, err := toml.Decode(text, &result)
	return &result, err
}

func Load() (*configdomain.PartialConfig, error) {
	file, err := os.Open(FileName)
	if err != nil {
		return nil, nil //nolint:nilerr,nilnil
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf(messages.ConfigFileCannotRead, ".git-branches.yml", err)
	}
	configFileData, err := Decode(string(bytes))
	if err != nil {
		return nil, fmt.Errorf(messages.ConfigFileInvalidData, ".git-branches.yml", err)
	}
	result, err := Validate(*configFileData)
	return &result, err
}

// Validate converts the given low-level configfile data into high-level config data.
func Validate(data Data) (configdomain.PartialConfig, error) {
	result := configdomain.PartialConfig{} //nolint:exhaustruct
	var err error
	if data.Branches != nil {
		if data.Branches.Main != nil {
			result.MainBranch = gitdomain.NewLocalBranchNameRef(*data.Branches.Main)
		}
		if data.Branches.Perennials != nil {
			result.PerennialBranches = gitdomain.NewLocalBranchNamesRef(data.Branches.Perennials...)
		}
	}
	if data.CodeHosting != nil {
		if data.CodeHosting.Platform != nil {
			result.CodeHostingPlatformName = configdomain.NewCodeHostingPlatformNameRef(*data.CodeHosting.Platform)
		}
		if data.CodeHosting.OriginHostname != nil {
			result.CodeHostingOriginHostname = configdomain.NewCodeHostingOriginHostnameRef(*data.CodeHosting.OriginHostname)
		}
	}
	if data.SyncStrategy != nil {
		if data.SyncStrategy.FeatureBranches != nil {
			result.SyncFeatureStrategy, err = configdomain.NewSyncFeatureStrategyRef(*data.SyncStrategy.FeatureBranches)
		}
		if data.SyncStrategy.PerennialBranches != nil {
			result.SyncPerennialStrategy, err = configdomain.NewSyncPerennialStrategyRef(*data.SyncStrategy.PerennialBranches)
		}
	}
	if data.PushNewbranches != nil {
		result.NewBranchPush = configdomain.NewNewBranchPushRef(*data.PushNewbranches)
	}
	if data.ShipDeleteTrackingBranch != nil {
		result.ShipDeleteTrackingBranch = configdomain.NewShipDeleteTrackingBranchRef(*data.ShipDeleteTrackingBranch)
	}
	if data.SyncBeforeShip != nil {
		result.SyncBeforeShip = configdomain.NewSyncBeforeShipRef(*data.SyncBeforeShip)
	}
	if data.SyncUpstream != nil {
		result.SyncUpstream = configdomain.NewSyncUpstreamRef(*data.SyncUpstream)
	}
	return result, err
}
