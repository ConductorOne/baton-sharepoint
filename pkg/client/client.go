package client

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

// shamelessly copied from https://github.com/ConductorOne/baton-microsoft-entra/

const (
	apiVersion  = "v1.0"
	betaVersion = "beta"
)

// makeGraphReadScopes is a helper function that generates a default graph scope.
// documentation: https://learn.microsoft.com/en-us/entra/identity-platform/scopes-oidc#the-default-scope
func makeGraphReadScopes(graphDomain string) []string {
	if graphDomain == "" {
		graphDomain = "graph.microsoft.com"
	}

	return []string{fmt.Sprintf("https://%s/.default", graphDomain)}
}

type Client struct {
	GraphDomain string
	token       azcore.TokenCredential
	http        *uhttp.BaseHttpClient
}

type QueryOption func(*queryOptions)

type queryOptions struct {
	skipEventualConsistency bool
}

func WithoutEventualConsistency() QueryOption {
	return func(o *queryOptions) {
		o.skipEventualConsistency = true
	}
}

func (c *Client) buildURL(reqPath string, v url.Values) string {
	ux := url.URL{
		Scheme:   "https",
		Host:     c.GraphDomain,
		Path:     path.Join(apiVersion, reqPath),
		RawQuery: v.Encode(),
	}
	return ux.String()
}

func (c *Client) query(ctx context.Context, scopes []string, method, requestURL string, body, res any, opts ...QueryOption) error {
	qOpts := &queryOptions{}
	for _, opt := range opts {
		opt(qOpts)
	}

	uri, err := url.Parse(requestURL)
	if err != nil {
		return err
	}

	token, err := c.token.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: scopes,
	})
	if err != nil {
		return err
	}

	reqOptions := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(token.Token),
	}

	if !qOpts.skipEventualConsistency {
		reqOptions = append(reqOptions, uhttp.WithHeader("ConsistencyLevel", "eventual"))
	}

	if body != nil {
		reqOptions = append(reqOptions, uhttp.WithJSONBody(body))
	}

	req, err := c.http.NewRequest(ctx, method, uri, reqOptions...)
	if err != nil {
		return err
	}

	doOptions := []uhttp.DoOption{}
	if res != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(res))
	}

	resp, err := c.http.Do(req, doOptions...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func New(ctx context.Context, tenantID, clientID, clientSecret, graphDomain string) (*Client, error) {
	var cred azcore.TokenCredential

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

	options := azcore.ClientOptions{
		Transport: httpClient,
	}

	cred, err = azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, &azidentity.ClientSecretCredentialOptions{
		ClientOptions: options,
	})
	if err != nil {
		return nil, err
	}

	http, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		token:       cred,
		http:        http,
		GraphDomain: graphDomain,
	}, nil
}
