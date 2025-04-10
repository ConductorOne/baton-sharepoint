package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/pagination"
)

// ListSites List sites across geographies in an organization. This
// API can also be used to enumerate all sites in a non-multi-geo
// tenant.
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
