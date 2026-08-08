// Package process provides utilities for pre-processing and parameter management.
package process

import (
	"github.com/ognerezov/hot-core/aws"
)

const (
	GoogleClientId       = "google_client_id"
	GoogleClientSecret   = "google_client_secret"
	FacebookClientId     = "facebook_client_id"
	FacebookClientSecret = "facebook_client_secret"
	AppleClientId        = "apple_client_id"
	AppleKeyId           = "apple_key_id"
	AppleTeamId          = "apple_team_id"
)

// SecretNames is a default list of secret names to be pre-processed.
var SecretNames = []string{
	GoogleClientId,
	GoogleClientSecret,
	FacebookClientId,
	FacebookClientSecret,
	AppleClientId,
	AppleKeyId,
	AppleTeamId,
}

// PreProcess retrieves multiple secure parameters from AWS SSM and returns them as a map.
func PreProcess(paramNames []string) (map[string]*string, error) {
	res := make(map[string]*string)

	for _, name := range paramNames {
		val, err := aws.GetSecureParameter(name)
		if err != nil {
			return nil, err
		}
		res[name] = val
	}

	return res, nil
}
