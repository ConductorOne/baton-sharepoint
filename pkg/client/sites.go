package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/pagination"
)

// ListSites List sites across geographies in an organization. This
// API can also be used to enumerate all sites in a non-multi-geo
// tenant.
//
// Permission required: `Sites.Read.All`
// documentation: https://learn.microsoft.com/en-us/graph/api/site-getallsites
func (c *Client) ListSites(ctx context.Context, bag *pagination.Bag) ([]Site, error) {
	targetURL := c.buildURL("sites/getllSites", url.Values{})
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

// GetUserInformationListItems list users and groups of a List's «User Information List».
// The User Information List is a hidden list that exist in each SharePoint Site.
//
// Permission required: `Sites.Read.All`
// documentation (on User Information List): From Microsoft? nowhere to be found 🤷
// documentation: https://learn.microsoft.com/en-us/graph/api/list-get
func (c *Client) GetUserInformationListItems(ctx context.Context, siteID string) error {
	var defaultUrlValues url.Values
	defaultUrlValues.Add("$expand", "Fields")

	targetURL := c.buildURL(fmt.Sprintf("/sites/%s/lists/User Information List/items", siteID), defaultUrlValues)

	var resp GetUserInformationListItemsResponse
	err := c.query(ctx, makeGraphReadScopes(c.GraphDomain), http.MethodGet, targetURL, nil, &resp)
	if err != nil {
		return fmt.Errorf("GetUserInformationListItems: request failed, error: %w", err)
	}

	return nil
}
