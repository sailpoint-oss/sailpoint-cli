package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

type Message struct {
	Locale       string `json:"locale,omitempty"`
	LocaleOrigin string `json:"localeOrigin,omitempty"`
	Text         string `json:"text,omitempty"`
}

type SDKResp struct {
	DetailCode string        `json:"detailCode,omitempty"`
	TrackingID string        `json:"trackingId,omitempty"`
	Messages   []Message     `json:"messages,omitempty"`
	Causes     []interface{} `json:"causes,omitempty"`
}

func HandleSDKError(resp *http.Response, sdkErr error) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}

	var formattedBody SDKResp
	err = json.Unmarshal(body, &formattedBody)
	if err != nil {
		log.Error(err)
	}

	outputErr := fmt.Sprintf("%s\ndate: %s\nslpt-request-id: %s\nmsgs:\n", sdkErr, resp.Header["Date"][0], resp.Header["Slpt-Request-Id"][0])

	if len(formattedBody.Messages) > 0 {
		for _, v := range formattedBody.Messages {
			outputErr = outputErr + fmt.Sprintf("%s\n", v.Text)
		}
	} else if len(body) > 0 {
		outputErr = outputErr + fmt.Sprintf("%s\n", string(body))
	}

	return errors.New(outputErr)

}
