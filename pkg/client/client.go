package client

import (
	"context"
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

	microsoftBuiltinAppsOwnerID = "f8cdef31-a31e-4b4a-93e4-5f571e91255a"

	// https://learn.microsoft.com/en-us/graph/api/resources/approleassignment?view=graph-rest-1.0
	//
	//  	The identifier (id) for the app role which is assigned
	//	to the principal. This app role must be exposed in the
	//	appRoles property on the resource application's
	//	service principal (resourceId). If the resource
	//	application has not declared any app roles, a default
	//	app role ID of 00000000-0000-0000-0000-000000000000
	//	can be specified to signal that the principal is
	//	assigned to the resource app without any specific app
	//	roles. Required on create
	//
	defaultAppRoleAssignmentID = "00000000-0000-0000-0000-000000000000"
)

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

func (c *Client) buildBetaURL(reqPath string, v url.Values) string {
	ux := url.URL{
		Scheme:   "https",
		Host:     c.GraphDomain,
		Path:     path.Join(betaVersion, reqPath),
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

func New(ctx context.Context, useCliCredentials bool, tenantID, clientID, clientSecret, graphDomain string) (*Client, error) {
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

	switch {
	case useCliCredentials:
		cred, err = azidentity.NewAzureCLICredential(nil)
	case tenantID != "" && clientID != "" && clientSecret != "":
		cred, err = azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, &azidentity.ClientSecretCredentialOptions{
			ClientOptions: options,
		})
	default:
		cred, err = azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
			ClientOptions: options,
			TenantID:      tenantID,
		})
	}
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
