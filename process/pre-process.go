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
	return PreProcessInRegion("", paramNames)
}

// PreProcessInRegion retrieves multiple secure parameters from AWS SSM in the specified region.
func PreProcessInRegion(region string, paramNames []string) (map[string]*string, error) {
	res := make(map[string]*string)

	for _, name := range paramNames {
		val, err := aws.GetSecureParameterInRegion(region, name)
		if err != nil {
			return nil, err
		}
		res[name] = val
	}

	return res, nil
}

// PreProcessInRegions retrieves multiple secure parameters from AWS SSM across multiple regions
// and returns a map where keys are regions and values are maps of parameter names to parameter values.
func PreProcessInRegions(regions []string, paramNames []string) (map[string]map[string]*string, error) {
	res := make(map[string]map[string]*string)

	for _, region := range regions {
		params, err := PreProcessInRegion(region, paramNames)
		if err != nil {
			return nil, err
		}
		res[region] = params
	}

	return res, nil
}

// PreProcessRegions is an alias for PreProcessInRegions.
func PreProcessRegions(regions []string, paramNames []string) (map[string]map[string]*string, error) {
	return PreProcessInRegions(regions, paramNames)
}
