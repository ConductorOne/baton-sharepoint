package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	cert_based_bearer_token "github.com/conductorone/baton-sharepoint/pkg/client/cert-based-bearer-token"
)

func (c *Client) ListGroupsForSite(ctx context.Context, siteWebURL string) ([]SharePointSiteGroup, error) {
	bearer, err := c.spTokenClient.GetBearerToken(ctx, cert_based_bearer_token.JWTOptions{
		ClientID:   c.clientID,
		TenantID:   c.tenantID,
		TimeUTCNow: time.Now().UTC(),
		Duration:   1 * time.Hour,
		NotBefore:  5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch bearer token for SharePoint REST API, error: %w", err)
	}

	url, err := url.Parse(siteWebURL)
	if err != nil {
		return nil, err
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(bearer),
	}

	url.Path = path.Join(url.Path, "/_api/web/sitegroups")

	req, err := c.http.NewRequest(ctx, http.MethodGet, url, reqOpts...)
	if err != nil {
		return nil, err
	}

	var data ListGroupsForSiteResponse
	resp, err := c.http.Do(req, uhttp.WithResponse(&data))
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return data.Value, nil
}

func (c *Client) ListUsersInGroupByGroupID(ctx context.Context, siteWebURL string, groupID int) ([]SharePointSiteUser, error) {
	bearer, err := c.spTokenClient.GetBearerToken(ctx, cert_based_bearer_token.JWTOptions{
		ClientID:   c.clientID,
		TenantID:   c.tenantID,
		TimeUTCNow: time.Now().UTC(),
		Duration:   1 * time.Hour,
		NotBefore:  5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch bearer token for SharePoint REST API, error: %w", err)
	}

	url, err := url.Parse(siteWebURL)
	if err != nil {
		return nil, err
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(bearer),
	}

	// TODO(shackra): this may need pagination
	url.Path = path.Join(url.Path, fmt.Sprintf("_api/web/SiteGroups/GetById(%d)/Users", groupID))

	req, err := c.http.NewRequest(ctx, http.MethodGet, url, reqOpts...)
	if err != nil {
		return nil, err
	}

	var data ListUsersInGroupByGroupIDResponse
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data))
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return data.Value, nil
}
