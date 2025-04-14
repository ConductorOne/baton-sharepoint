package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/pagination"
)

// ListSites List sites in an organization.
//
// Permission required: `Sites.Read.All`
// documentation: https://learn.microsoft.com/en-us/graph/api/site-getallsites
func (c *Client) ListSites(ctx context.Context, bag *pagination.Bag) ([]Site, error) {
	defaultValues := url.Values{}
	defaultValues.Set("search", "")
	defaultValues.Set("$select", strings.Join([]string{"id", "name", "displayName", "isPersonalSite", "siteCollection", "webUrl", "root"}, ","))
	defaultValues.Set("$top", "999")

	targetURL := c.buildURL("sites", defaultValues)
	if bag.PageToken() != "" {
		targetURL = bag.PageToken()
	}

	var resp GetAllSitesResponse
	err := c.query(ctx, makeGraphReadScopes(c.GraphDomain), http.MethodGet, targetURL, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("ListSites: request failed, error: %w", err)
	}
	if resp.NextLink != "" {
		err := bag.Next(resp.NextLink)
		if err != nil {
			return nil, fmt.Errorf("ListSites: pagination: cannot set next page token, error: %w", err)
		}
	}

	return resp.Value, nil
}

// GetSiteByID fetch a sites.
//
// Permission required: `Sites.Read.All`
// documentation: https://learn.microsoft.com/es-es/graph/api/site-get
func (c *Client) GetSiteByID(ctx context.Context, id string) (*Site, error) {
	defaultValues := url.Values{}
	defaultValues.Set("$select", strings.Join([]string{"id", "name", "displayName", "siteCollection", "webUrl", "root"}, ","))

	targetURL := c.buildURL(path.Join("sites", id), defaultValues)
	var resp Site

	err := c.query(ctx, makeGraphReadScopes(c.GraphDomain), http.MethodGet, targetURL, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("GetSiteByID: request failed, error: %w", err)
	}

	return &resp, nil
}
