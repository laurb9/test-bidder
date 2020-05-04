package bidder

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mxmCherry/openrtb/openrtb2"
)

var debug = os.Getenv("DEBUG")

// TestParams schema for parameters for controlling test bidder responses
type TestParams struct {
	// Delay for tmax milliseconds before returning the response
	TMax int `json:"tmax,omitempty"`
}

// ReadTestParams reads bidder params from site/app.ext (because the adapter may not pass ext entirely)
func ReadTestParams(bidRequest *openrtb2.BidRequest) (params TestParams) {
	// TODO: have a schema for this
	brExt := bidRequest.Ext
	if err := json.Unmarshal(brExt, &params); err != nil {
		log.Printf("ext json decode error: %s", err.Error())
	}
	if debug == "1" {
		log.Printf("using params %+v", params)
	}
	return params
}

// Handler is an OpenRTB 2.5 test bidder
func Handler(w http.ResponseWriter, r *http.Request) {
	var (
		ctx         context.Context
		cancel      context.CancelFunc
		bidResponse *openrtb2.BidResponse
		bidRequest  *openrtb2.BidRequest
	)
	start := time.Now()

	if debug == "1" {
		log.Printf("RequestLen=%d, Headers=%v", r.ContentLength, r.Header)
	}

	bidRequest = new(openrtb2.BidRequest)
	if err := json.NewDecoder(r.Body).Decode(bidRequest); err != nil {
		log.Printf("json decode error: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("json decode error"))
		return
	}

	params := ReadTestParams(bidRequest)

	// Create a context
	timeLeft := time.Duration(bidRequest.TMax)*time.Millisecond - time.Since(start)
	ctx, cancel = context.WithTimeout(r.Context(), timeLeft)
	defer cancel()

	if len(bidRequest.Imp) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("imp is empty"))
		return
	}
	imp1 := bidRequest.Imp[0]

	// Prepare a bid response
	// Additional extra data for bid1
	ext1, _ := json.Marshal(map[string]interface{}{})
	// Bid 1
	bid1 := openrtb2.Bid{
		ID:      "test-bid-id-1",
		ImpID:   imp1.ID,
		Price:   0.1,
		CrID:    "test-creative-id-1",
		W:       720,
		H:       80,
		AdM:     `<html><a href="//localhost"><img src="//localhost/ad.img"></a></html>`,
		AdID:    "test-ad-id-12345",
		ADomain: []string{"example.com"},
		Ext:     ext1,
	}

	// Customize bid by type
	switch {
	case imp1.Banner != nil:
		if formats := imp1.Banner.Format; len(formats) > 0 {
			bid1.W = formats[0].W
			bid1.H = formats[0].H
		}

	case imp1.Video != nil:
		if bid1.W, bid1.H = imp1.Video.W, imp1.Video.H; bid1.W == 0 || bid1.H == 0 {
			bid1.W = bidRequest.Device.W
			bid1.H = bidRequest.Device.H
		}
		bid1.AdM = `<?xml version="1.0" encoding="UTF-8"?>
<VAST version="2.0">
	<Ad id="1">
		<Wrapper>
			<AdSystem>OX</AdSystem>
			<VASTAdTagURI><![CDATA[https://localhost/vast.xml]]></VASTAdTagURI>
			<Impression></Impression>
			<Creatives>
				<Creative>
					<Linear>
						<TrackingEvents>
							<ClickTracking><![CDATA[https://localhost/click]]></ClickTracking>
						</TrackingEvents>
						<VideoClicks></VideoClicks>
					</Linear>
				</Creative>
			</Creatives>
		</Wrapper>
	</Ad>
</VAST>`

	case imp1.Audio != nil:
	case imp1.Native != nil:
	}

	bidResponse = &openrtb2.BidResponse{
		ID: bidRequest.ID,
		SeatBid: []openrtb2.SeatBid{
			{
				Seat:  "Bidder",
				Group: 0,
				Bid: []openrtb2.Bid{
					bid1,
				},
			},
		},
		BidID:      "TEST_BID_ID",
		Cur:        "EUR",
		CustomData: "",
		NBR:        nil,
		Ext:        json.RawMessage(`{"pid": "bar"}`),
	}

	if params.TMax > 0 {
		responseTime := time.Duration(float64(params.TMax))*time.Millisecond - time.Since(start)
		wait := time.After(responseTime)
		select {
		case <-ctx.Done():
			log.Printf("bidRequest.TMAX=%dms expired before requested tmax=%dms", bidRequest.TMax, responseTime.Milliseconds())
		case <-wait:
		}
	}

	if content, err := json.Marshal(bidResponse); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	} else {
		log.Printf("Error marshalling response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	ctx.Done()
}
