package template

import (
	"log"

	"github.com/mxmCherry/openrtb/openrtb2"
)

// TestParams schema for parameters for controlling test bidder responses
type TestParams struct {
	// Delay for tmax milliseconds before returning the response
	TMax int `json:"tmax,omitempty"`
}

// ReadTestParams reads bidder params from site/app.ext (because the adapter may not pass ext entirely)
func ReadTestParams(bidRequest *openrtb2.BidRequest) (params TestParams) {
	if debug == "1" {
		log.Printf("using params %+v", params)
	}
	return params
}
