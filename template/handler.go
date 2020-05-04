package template

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

// Handler is the OpenRTB 2.5 templated test bidder
func Handler(w http.ResponseWriter, r *http.Request) {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		bidRequest *openrtb2.BidRequest
	)
	start := time.Now()
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

	tmpl := NewTemplate("response-templates", "bid-response.json")
	if debug == "1" {
		log.Printf(tmpl.String())
	}

	w.Header().Set("Content-Type", "application/json")

	// TODO: re-parse output as BidResponse to avoid sending bad or incomplete json response
	if err := tmpl.tpl.Execute(w, bidRequest); err != nil {
		log.Printf("Template execution failed: %v", err)
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
	ctx.Done()
}
