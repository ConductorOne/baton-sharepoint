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

func (c *Client) ListUsersInGroupByGroupID(ctx context.Context, groupURLID string) ([]SharePointUser, error) {
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

func (c *Client) ListSharePointUsers(ctx context.Context, siteWebURL string) ([]SharePointUser, error) {
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

func (c *Client) RemoveThingFromGroupByThingID(ctx context.Context, siteWebURL string, groupID, thingID int) error {
	bearer, err := c.certbasedToken.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{fmt.Sprintf(scopeSharePointTemplate, c.sharePointDomain)},
	})
	if err != nil {
		return fmt.Errorf("Client.ListSharePointUsers: failed to fetch bearer token, error: %w", err)
	}

	site, err := GuessSharePointSiteWebURLBase(siteWebURL)
	if err != nil {
		return err
	}

	url, err := url.Parse(site)
	if err != nil {
		return err
	}

	digest, err := c.getFormDigestValue(ctx, site)
	if err != nil {
		return err
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeFormHeader(),
		uhttp.WithBearerToken(bearer.Token),
		uhttp.WithHeader("X-RequestDigest", digest),
		uhttp.WithHeader("X-HTTP-Method", "Delete"),
		uhttp.WithFormBody(""),
	}

	url.Path = path.Join(url.Path, fmt.Sprintf("_api/web/sitegroups(%d)/users/removebyid(%d)", groupID, thingID))

	req, err := c.http.NewRequest(ctx, http.MethodPost, url, reqOpts...)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}


func (c *Client) AddThingToGroupByThingID(ctx context.Context, siteWebURL string, groupID int, thingID string) (*SharePointUser, error) {
	bearer, err := c.certbasedToken.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{fmt.Sprintf(scopeSharePointTemplate, c.sharePointDomain)},
	})
	if err != nil {
		return nil, fmt.Errorf("Client.ListSharePointUsers: failed to fetch bearer token, error: %w", err)
	}

	site, err := GuessSharePointSiteWebURLBase(siteWebURL)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(site)
	if err != nil {
		return nil, err
	}

	loginName := guessFullLoginName(thingID)

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithBearerToken(bearer.Token),
		uhttp.WithJSONBody(&SharePointAddThingRequest{
			Metadata:  SharePointAddThingMetadata{Type: "SP.User"},
			LoginName: loginName,
		}),
		uhttp.WithHeader("Content-Type", "application/json;odata=verbose"),
	}

	url.Path = path.Join(url.Path, fmt.Sprintf("_api/web/sitegroups(%d)/users", groupID))

	req, err := c.http.NewRequest(ctx, http.MethodPost, url, reqOpts...)
	if err != nil {
		return nil, err
	}

	var data SharePointUser
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data))
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return &data, nil
}

// EnsureUserByUserPrincipalName ask the SharePoint site to give an ID to a user by its User Principal Name.
func (c *Client) EnsureUserByUserPrincipalName(ctx context.Context, siteWebURL, logonName string) (*SharePointUser, error) {
	bearer, err := c.certbasedToken.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{fmt.Sprintf(scopeSharePointTemplate, c.sharePointDomain)},
	})
	if err != nil {
		return nil, fmt.Errorf("Client.EnsureUserByUserPrincipalName: failed to fetch bearer token, error: %w", err)
	}

	site, err := GuessSharePointSiteWebURLBase(siteWebURL)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(site)
	if err != nil {
		return nil, err
	}

	thing := strings.Join(append(userLoginName, logonName), "|")

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithBearerToken(bearer.Token),
		uhttp.WithJSONBody(&SharePointEnsureThingRequest{
			LogonName: thing,
		}),
		uhttp.WithHeader("Content-Type", "application/json;odata=verbose"),
	}

	u.Path = path.Join(u.Path, "_api/web/ensureuser")

	req, err := c.http.NewRequest(ctx, http.MethodPost, u, reqOpts...)
	if err != nil {
		return nil, err
	}

	var data SharePointUser
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data))
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return &data, nil
}

// EnsureUserByUserID ask the SharePoint site to give an ID to a user.
func (c *Client) EnsureUserByUserID(ctx context.Context, siteWebURL, userID string) (*SharePointUser, error) {
	user, err := c.GetUserPrincipalNameFromUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Client.EnsureUserByUserID: failed to get user principal name for user ID '%s', error: %w", userID, err)
	}

	return c.EnsureUserByUserPrincipalName(ctx, siteWebURL, user)
}
