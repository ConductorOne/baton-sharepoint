package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/pagination"
)

// GetUserInformationListItems list users and groups of a List's «User Information List».
// The User Information List is a hidden list that exist in each SharePoint Site.
//
// Permission required: `Sites.Read.All`
// documentation (on User Information List): From Microsoft? nowhere to be found 🤷
// documentation: https://learn.microsoft.com/en-us/graph/api/list-get
func (c *Client) GetUserInformationListItems(ctx context.Context, siteID string, bag *pagination.Bag) error {
	var defaultUrlValues url.Values
	defaultUrlValues.Set("$expand", "Fields")
	defaultUrlValues.Set("$top", "999")

	targetURL := c.buildURL(fmt.Sprintf("/sites/%s/lists/User Information List/items", siteID), defaultUrlValues)
	if bag.PageToken() != "" {
		targetURL = bag.PageToken()
	}

	var resp GetUserInformationListItemsResponse
	err := c.query(ctx, makeGraphReadScopes(c.GraphDomain), http.MethodGet, targetURL, nil, &resp)
	if err != nil {
		return fmt.Errorf("GetUserInformationListItems: request failed, error: %w", err)
	}
	if resp.NextLink != "" {
		err := bag.Next(resp.NextLink)
		if err != nil {
			return fmt.Errorf("GetUserInformationListItems: pagination: cannot set next page token, error: %w", err)
		}
	}

	return nil
}
