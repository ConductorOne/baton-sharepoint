package cert_based_bearer_token

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

var (
	oauthEndpointTemplate   = "https://login.microsoftonline.com/$tenantId/oauth2/v2.0/token"
	scopeSharePointTemplate = "https://%s.sharepoint.com/.default"
)

type Exchange struct {
	cacheBearerToken    string
	cacheBearerTokenExp time.Time
	SharePointDomain    string

	http *uhttp.BaseHttpClient
}

func (e *Exchange) GetBearerToken(ctx context.Context, opts JWTOptions) (string, error) {
	// return the cache token we got from a previous exchange
	if !e.cacheBearerTokenExp.IsZero() && opts.TimeUTCNow.Before(e.cacheBearerTokenExp) {
		return e.cacheBearerToken, nil
	}

	uri, err := url.Parse(fmt.Sprintf(oauthEndpointTemplate, opts.TenantID))
	if err != nil {
		return "", err
	}

	jwt, err := generateSignedJWTFromPFX(opts)
	if err != nil {
		return "", fmt.Errorf("cannot generate JWT for exchange, error: %w", err)
	}

	body := url.Values{
		"client_id":             []string{opts.ClientID},
		"scope":                 []string{fmt.Sprintf(scopeSharePointTemplate, e.SharePointDomain)},
		"grant_type":            []string{"client_credentials"},
		"client_assertion_type": []string{"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      []string{jwt},
	}

	res := &MicrosoftOAuthResponse{}

	reqOptions := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeFormHeader(),
		uhttp.WithFormBody(body.Encode()),
	}

	req, err := e.http.NewRequest(ctx, http.MethodPost, uri, reqOptions...)
	if err != nil {
		return "", err
	}

	resp, err := e.http.Do(req, uhttp.WithJSONResponse(res))
	if err != nil {
		return "", err
	}

	// TODO(shackra): check the StatusCode?
	resp.Body.Close()

	if res.AccessToken != "" {
		e.cacheBearerToken = res.AccessToken
		e.cacheBearerTokenExp = opts.TimeUTCNow.Add(opts.Duration)
	} else {
		return "", fmt.Errorf("Something is wrong, Microsoft has responded with an empty token!")
	}

	return res.AccessToken, nil
}
