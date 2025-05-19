package client

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"software.sslmate.com/src/go-pkcs12"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-sharepoint/pkg/errorexplained"
)

// shamelessly copied from https://github.com/ConductorOne/baton-microsoft-entra/

const (
	apiVersion              = "v1.0"
	betaVersion             = "beta"
	scopeSharePointTemplate = "https://%s.sharepoint.com/.default"
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
	GraphDomain    string
	token          azcore.TokenCredential
	certbasedToken azcore.TokenCredential
	http           *uhttp.BaseHttpClient

	// SharePoint related stuff
	tenantID         string
	clientID         string
	sharePointDomain string

	// SharePointHome OrgLinks groups related stuff
	//
	// if this is set to true, the costumer needs to grant the permission
	// SharePoint > Sites.FullControl.All; that's basically full admin
	// rights over all SharePoint sites!
	dontFilterSharePointSpecialGroups bool
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

	var queryErr errorexplained.ErrorExplained
	doOptions := []uhttp.DoOption{
		uhttp.WithErrorResponse(&queryErr),
	}
	if res != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(res))
	}

	resp, err := c.http.Do(req, doOptions...)
	if err != nil {
		return errorexplained.WhatErrorToReturn(queryErr, err, "")
	}

	defer resp.Body.Close()

	return nil
}

func New(ctx context.Context, tenantID, clientID, clientSecret, graphDomain, sharepointDomain, pfxCert, pfxCertPassword string, syncSharePointHomeOrgLinks bool) (*Client, error) {
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

	// we cannot use errorexplained package with `cred`, azidentity has full control of the authentication flow
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, &azidentity.ClientSecretCredentialOptions{
		ClientOptions: options,
	})
	if err != nil {
		return nil, err
	}

	var pfxData []byte
	// Check if the certificate is provided as a base64-encoded string
	if _, err := base64.StdEncoding.DecodeString(pfxCert); err == nil {
		pfxData, err = base64.StdEncoding.DecodeString(pfxCert)
		if err != nil {
			return nil, fmt.Errorf("cannot decode base64 string of .pfx certificate, error: %w", err)
		}
	} else {
		// If not a valid base64 string, use the raw content as is (it was loaded from a file)
		pfxData = []byte(pfxCert)
	}

	privkey, cert, err := pkcs12.Decode(pfxData, pfxCertPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt .pfx certificate with password, error: %w", err)
	}

	rsaKey, ok := privkey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}

	certcred, err := azidentity.NewClientCertificateCredential(
		tenantID,
		clientID,
		[]*x509.Certificate{cert},
		rsaKey,
		&azidentity.ClientCertificateCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Transport: httpClient,
			},
			SendCertificateChain: true,
		},
	)
	if err != nil {
		return nil, err
	}

	http, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		token:                             cred,
		certbasedToken:                    certcred,
		http:                              http,
		GraphDomain:                       graphDomain,
		tenantID:                          tenantID,
		clientID:                          clientID,
		sharePointDomain:                  sharepointDomain,
		dontFilterSharePointSpecialGroups: syncSharePointHomeOrgLinks,
	}, nil
}
