/*
IdentityNow V3 API

Use these APIs to interact with the IdentityNow platform to achieve repeatable, automated processes with greater scalability. We encourage you to join the SailPoint Developer Community forum at https://developer.sailpoint.com/discuss to connect with other developers using our APIs.

API version: 3.0.0
*/

package sailpoint

import (
	"regexp"

	"github.com/hashicorp/go-retryablehttp"
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/sdk-output/beta"
	sailpointccsdk "github.com/sailpoint-oss/golang-sdk/sdk-output/cc"
	sailpointv2sdk "github.com/sailpoint-oss/golang-sdk/sdk-output/v2"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/sdk-output/v3"
)

var (
	jsonCheck = regexp.MustCompile(`(?i:(?:application|text)/(?:vnd\.[^;]+\+)?json)`)
	xmlCheck  = regexp.MustCompile(`(?i:(?:application|text)/xml)`)
)

// APIClient manages communication with the IdentityNow V3 API API v3.0.0
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	cfg    *sailpointsdk.Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// API Services

	V3    *sailpointsdk.APIClient
	V2    *sailpointv2sdk.APIClient
	Beta  *sailpointbetasdk.APIClient
	CC    *sailpointccsdk.APIClient
	token string
}

type service struct {
	client     *sailpointsdk.APIClient
	v2client   *sailpointv2sdk.APIClient
	betaClient *sailpointbetasdk.APIClient
	ccClient   *sailpointccsdk.APIClient
}

// NewAPIClient creates a new API client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(cfg *Configuration) *APIClient {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = retryablehttp.NewClient()
	}

	c := &APIClient{}

	CV3 := sailpointsdk.NewConfiguration(cfg.ClientConfiguration.ClientId, cfg.ClientConfiguration.ClientSecret, cfg.ClientConfiguration.BaseURL+"/v3", cfg.ClientConfiguration.TokenURL, cfg.ClientConfiguration.Token)
	//CV2 := sailpointv2sdk.NewConfiguration(cfg.ClientConfiguration.ClientId, cfg.ClientConfiguration.ClientSecret, cfg.ClientConfiguration.BaseURL+"/v2", cfg.ClientConfiguration.TokenURL)
	CBeta := sailpointbetasdk.NewConfiguration(cfg.ClientConfiguration.ClientId, cfg.ClientConfiguration.ClientSecret, cfg.ClientConfiguration.BaseURL+"/beta", cfg.ClientConfiguration.TokenURL, cfg.ClientConfiguration.Token)
	//CCC := sailpointccsdk.NewConfiguration(cfg.ClientConfiguration.ClientId, cfg.ClientConfiguration.ClientSecret, cfg.ClientConfiguration.BaseURL, cfg.ClientConfiguration.TokenURL)

	CV3.HTTPClient = cfg.HTTPClient
	//CV2.HTTPClient = cfg.HTTPClient
	CBeta.HTTPClient = cfg.HTTPClient
	//CCC.HTTPClient = cfg.HTTPClient

	c.V3 = sailpointsdk.NewAPIClient(CV3)
	c.V2 = sailpointv2sdk.NewAPIClient(sailpointv2sdk.NewConfiguration(cfg.ClientConfiguration.ClientId, cfg.ClientConfiguration.ClientSecret, cfg.ClientConfiguration.BaseURL+"/v2", cfg.ClientConfiguration.TokenURL))
	c.Beta = sailpointbetasdk.NewAPIClient(CBeta)
	c.CC = sailpointccsdk.NewAPIClient(sailpointccsdk.NewConfiguration(cfg.ClientConfiguration.ClientId, cfg.ClientConfiguration.ClientSecret, cfg.ClientConfiguration.BaseURL, cfg.ClientConfiguration.TokenURL))

	// API Services

	return c
}
