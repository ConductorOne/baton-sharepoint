package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-sharepoint/pkg/errorexplained"
)

// Documentation: https://learn.microsoft.com/en-us/previous-versions/office/developer/sharepoint-rest-reference/dn531432(v=office.15)

// NOTE(shackra): SharePoint REST API has no support for server-side pagination except, maybe, for List and List Items. The alternative is client-side
//                pagination.

func (c *Client) ListGroupsForSite(ctx context.Context, siteWebURL string) ([]SharePointSiteGroup, error) {
	bearer, err := c.certbasedToken.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{fmt.Sprintf(scopeSharePointTemplate, c.sharePointDomain)},
	})
	if err != nil {
		return nil, fmt.Errorf("Client.ListGroupsForSite: failed to fetch bearer token, error: %w", err)
	}

	url, err := url.Parse(siteWebURL)
	if err != nil {
		return nil, err
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(bearer.Token),
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

	if c.dontFilterSharePointSpecialGroups {
		return data.Value, nil
	}

	filtered := slices.DeleteFunc(data.Value, func(spg SharePointSiteGroup) bool {
		return strings.HasPrefix(spg.Title, "SharePointHome OrgLinks")
	})

	return filtered, nil
}

func (c *Client) ListSecurityPrincipalsInGroupByGroupID(ctx context.Context, groupURLID string) ([]SecurityPrincipal, error) {
	bearer, err := c.certbasedToken.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{fmt.Sprintf(scopeSharePointTemplate, c.sharePointDomain)},
	})
	if err != nil {
		return nil, fmt.Errorf("Client.ListUsersInGroupByGroupID: failed to fetch bearer token, error: %w", err)
	}

	url, err := url.Parse(groupURLID)
	if err != nil {
		return nil, err
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(bearer.Token),
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
		if strings.Contains(err.Error(), "403 Forbidden") && !c.dontFilterSharePointSpecialGroups {
			altMessage = fmt.Sprintf("access to the user list of group '%s' was denied, are we trying to list users of a 'special' group?", groupURLID)
		} else if strings.Contains(err.Error(), "403 Forbidden") && c.dontFilterSharePointSpecialGroups {
			altMessage = fmt.Sprintf("access to the user list of group '%s' was denied, check that admin consent was "+
				"granted for API permission SharePoint > Sites.FullControl.All for your registered app", groupURLID)
		}
		return nil, errorexplained.WhatErrorToReturn(queryErr, err, altMessage)
	}

	resp.Body.Close()

	return data.Value, nil
}

func (c *Client) ListSecurityPrincipals(ctx context.Context, siteWebURL string) ([]SecurityPrincipal, error) {
	bearer, err := c.certbasedToken.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{fmt.Sprintf(scopeSharePointTemplate, c.sharePointDomain)},
	})
	if err != nil {
		return nil, fmt.Errorf("Client.ListSharePointUsers: failed to fetch bearer token, error: %w", err)
	}

	url, err := url.Parse(siteWebURL)
	if err != nil {
		return nil, err
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(bearer.Token),
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
