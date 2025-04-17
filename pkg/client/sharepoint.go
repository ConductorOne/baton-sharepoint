package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	cbbt "github.com/conductorone/baton-sharepoint/pkg/client/cert-based-bearer-token"
)

// Documentation: https://learn.microsoft.com/en-us/previous-versions/office/developer/sharepoint-rest-reference/dn531432(v=office.15)

func (c *Client) ListGroupsForSite(ctx context.Context, siteWebURL string) ([]SharePointSiteGroup, error) {
	bearer, err := c.spTokenClient.GetBearerToken(ctx, cbbt.JWTOptions{
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

func (c *Client) ListUsersInGroupByGroupID(ctx context.Context, groupURLID string) ([]SharePointSiteUser, error) {
	bearer, err := c.spTokenClient.GetBearerToken(ctx, cbbt.JWTOptions{
		ClientID:   c.clientID,
		TenantID:   c.tenantID,
		TimeUTCNow: time.Now().UTC(),
		Duration:   1 * time.Hour,
		NotBefore:  5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch bearer token for SharePoint REST API, error: %w", err)
	}

	url, err := url.Parse(groupURLID)
	if err != nil {
		return nil, err
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(bearer),
	}

	// TODO(shackra): this may need pagination
	url.Path = path.Join(url.Path, "Users")

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

func (c *Client) ListSharePointUsers(ctx context.Context, siteWebURL string) ([]SharePointUser, error) {
	bearer, err := c.spTokenClient.GetBearerToken(ctx, cbbt.JWTOptions{
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
	url.Path = path.Join(url.Path, "_api/web/siteusers")

	req, err := c.http.NewRequest(ctx, http.MethodGet, url, reqOpts...)
	if err != nil {
		return nil, err
	}

	var data ListUsersResponse
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data))
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return data.Value, nil
}
