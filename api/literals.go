package api

import "github.com/ognerezov/hot-core/tools"

// SupportedRegions is the single source of truth for all active AWS deployment regions.
var SupportedRegions = []string{
	"us-east-1",
	"us-east-2",
	"eu-south-2",
	"il-central-1",
}

var CentralRegion = "us-east-1"

var ReplicationRegions = tools.RemoveAllStrings(AllRegions, CentralRegion)

var AllRegions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	"af-south-1",
	"ap-east-1",
	"ap-south-1",
	"ap-south-2",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-northeast-3",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-southeast-4",
	"ap-southeast-5",
	"ap-southeast-6",
	"ap-southeast-7",
	"ca-central-1",
	"ca-west-1",
	"eu-central-1",
	"eu-central-2",
	"eu-west-1",
	"eu-west-2",
	"eu-west-3",
	"eu-south-1",
	"eu-south-2",
	"eu-north-1",
	"il-central-1",
	"me-south-1",
	"me-central-1",
	"mx-central-1",
	"sa-east-1",
	"us-gov-east-1",
	"us-gov-west-1",
	"cn-north-1",
	"cn-northwest-1",
}

const (
	DefaultRegion         = "eu-south-2"
	CDKBootstrapStackName = "CDKToolkit"
)

// IsRegionSupported checks if the given region is present in the supported list.
func IsRegionSupported(region string) bool {
	for _, r := range SupportedRegions {
		if r == region {
			return true
		}
	}
	return false
}
