package template

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mxmCherry/openrtb/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestSimpleResponse(t *testing.T) {
	bidRequest := openrtb2.BidRequest{
		ID:   "1234",
		TMax: 1500,
		Imp: []openrtb2.Imp{
			{
				ID: "123456",
			},
		},
	}

	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(bidRequest)

	req := httptest.NewRequest("POST", "http://localhost/t/openrtb/2.5", buf)
	w := httptest.NewRecorder()

	Handler(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("Wrong status %d received, expected %d", w.Code, http.StatusOK)
	}

	resp := w.Result()

	bidResponse := new(openrtb2.BidResponse)
	json.NewDecoder(resp.Body).Decode(bidResponse)

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "1234", bidResponse.ID)
}
