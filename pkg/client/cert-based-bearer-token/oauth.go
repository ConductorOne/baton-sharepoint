package certbasedbearertoken

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

var (
	oauthEndpointTemplate   = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
	scopeSharePointTemplate = "https://%s.sharepoint.com/.default"
)

type Exchange struct {
	cacheBearerToken    string
	cacheBearerTokenExp time.Time
	sharePointDomain    string
	cert                string
	certPassword        string

	http *uhttp.BaseHttpClient
}

func (e *Exchange) GetBearerToken(ctx context.Context, opts JWTOptions) (string, error) {
	// return the cache token we got from a previous exchange
	if !e.cacheBearerTokenExp.IsZero() && opts.TimeUTCNow.Before(e.cacheBearerTokenExp) {
		return e.cacheBearerToken, nil
	}
	// fill the missing information
	opts.pfxBase64 = e.cert
	opts.password = e.certPassword

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
		"scope":                 []string{fmt.Sprintf(scopeSharePointTemplate, e.sharePointDomain)},
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
		return "", fmt.Errorf("something went wrong, Microsoft has responded with an empty token")
	}

	return res.AccessToken, nil
}

func New(ctx context.Context, sharepointdomain, cert, certPassword string) (*Exchange, error) {
	uhttpOptions := []uhttp.Option{
		uhttp.WithLogger(true, ctxzap.Extract(ctx)),
	}
	httpClient, err := uhttp.NewClient(
		ctx,
		uhttpOptions...,
	)
	if err != nil {
		return nil, err
	}

	http, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return &Exchange{http: http, sharePointDomain: sharepointdomain, cert: cert, certPassword: certPassword}, nil
}
