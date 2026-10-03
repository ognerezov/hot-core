// Package process provides utilities for pre-processing and parameter management.
package process

import (
	"fmt"

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

	RegionalBillingApiUrlParam = "regional_billing_api_url"
	AiRequestApiUrlParam       = "regional_ai_request_api_url"
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

var EmptyParamsToPopulate = []string{
	RegionalBillingApiUrlParam,
	AiRequestApiUrlParam,
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

// CopyInRegion copies a specified slice of parameters from srcRegion to dstRegion.
func CopyInRegion(srcRegion, dstRegion string, paramNames []string) error {
	if srcRegion == dstRegion {
		return nil
	}
	return aws.CopyParametersInRegion(srcRegion, dstRegion, paramNames)
}

// CopySecretNamesInRegion copies the default SecretNames list from srcRegion to dstRegion.
func CopySecretNamesInRegion(srcRegion, dstRegion string) error {
	return CopyInRegion(srcRegion, dstRegion, SecretNames)
}

// CopyInRegions copies a specified slice of parameters from srcRegion to multiple dstRegions.
func CopyInRegions(srcRegion string, dstRegions []string, paramNames []string) error {
	for _, dstRegion := range dstRegions {
		if err := CopyInRegion(srcRegion, dstRegion, paramNames); err != nil {
			return fmt.Errorf("failed to sync params to region %s: %w", dstRegion, err)
		}
	}
	return nil
}

// CopySecretNamesInRegions copies default SecretNames from srcRegion to multiple dstRegions.
func CopySecretNamesInRegions(srcRegion string, dstRegions []string) error {
	return CopyInRegions(srcRegion, dstRegions, SecretNames)
}

func BootStrapParamsInRegion(dstRegion string, params []string) error {
	for _, param := range params {
		if err := aws.PutParameterInRegion(dstRegion, param, ""); err != nil {
			return fmt.Errorf("failed to bootstrap param %s in region %s: %w", param, dstRegion, err)
		}
	}
	return nil
}

func BootStrapEmptyParamsInRegion(dstRegion string) error {
	return BootStrapParamsInRegion(dstRegion, EmptyParamsToPopulate)
}
