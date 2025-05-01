package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	cbbt "github.com/conductorone/baton-sharepoint/pkg/client/cert-based-bearer-token"
	"github.com/conductorone/baton-sharepoint/pkg/errorexplained"
)

// Documentation: https://learn.microsoft.com/en-us/previous-versions/office/developer/sharepoint-rest-reference/dn531432(v=office.15)

// NOTE(shackra): SharePoint REST API has no support for server-side pagination except, maybe, for List and List Items. The alternative is client-side
//                pagination.

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
	var queryErr errorexplained.ErrorExplained
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data), uhttp.WithErrorResponse(&queryErr))
	if err != nil {
		return nil, errorexplained.WhatErrorToReturn(queryErr, err, "")
	}

	resp.Body.Close()
	filtered := slices.DeleteFunc(data.Value, func(spg SharePointSiteGroup) bool {
		return strings.HasPrefix(spg.Title, "SharePointHome OrgLinks")
	})

	return filtered, nil
}

func (c *Client) ListUsersInGroupByGroupID(ctx context.Context, groupURLID string) ([]SharePointUser, error) {
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

	url.Path = path.Join(url.Path, "Users")

	req, err := c.http.NewRequest(ctx, http.MethodGet, url, reqOpts...)
	if err != nil {
		return nil, err
	}

	var data ListUsersInGroupByGroupIDResponse
	var queryErr errorexplained.ErrorExplained
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data), uhttp.WithErrorResponse(&queryErr))
	if err != nil {
		altMessage := ""
		if strings.Contains(err.Error(), "403 Forbidden") {
			altMessage = fmt.Sprintf("access to the user list of group '%s' was denied, are we trying to list users of a 'special' group?", groupURLID)
		}
		return nil, errorexplained.WhatErrorToReturn(queryErr, err, altMessage)
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

	url.Path = path.Join(url.Path, "_api/web/siteusers")

	req, err := c.http.NewRequest(ctx, http.MethodGet, url, reqOpts...)
	if err != nil {
		return nil, err
	}

	var data ListUsersResponse
	var queryErr errorexplained.ErrorExplained
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data), uhttp.WithErrorResponse(&queryErr))
	if err != nil {
		altMessage := ""
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			altMessage = ("cannot list SharePoint site users, check that admin consent was " +
				"granted for API permission SharePoint > User.Read.All for your registered app")
		}
		return nil, errorexplained.WhatErrorToReturn(queryErr, err, altMessage)
	}

	resp.Body.Close()

	return data.Value, nil
}
