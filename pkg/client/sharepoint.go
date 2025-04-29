package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	cbbt "github.com/conductorone/baton-sharepoint/pkg/client/cert-based-bearer-token"
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

func (c *Client) RemoveThingFromGroupByThingID(ctx context.Context, siteWebURL string, groupID, thingID int) error {
	bearer, err := c.spTokenClient.GetBearerToken(ctx, cbbt.JWTOptions{
		ClientID:   c.clientID,
		TenantID:   c.tenantID,
		TimeUTCNow: time.Now().UTC(),
		Duration:   1 * time.Hour,
		NotBefore:  5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("unable to fetch bearer token for SharePoint REST API, error: %w", err)
	}

	site, err := guessSharePointSiteWebURLBase(siteWebURL)
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
		uhttp.WithBearerToken(bearer),
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

var (
	userLoginName        = []string{"i:0#.f", "membership"}
	groupLoginName       = []string{"c:0o.c", "federateddirectoryclaimprovider"}
	rolemanagerLoginName = []string{"c:0-.f"}
	tenantLoginName      = []string{"c:0t.c"}
	allUsersWindows      = []string{"c:0!.s"}
)

func (c *Client) AddThingToGroupByThingID(ctx context.Context, siteWebURL string, groupID int, thingID string) error {
	bearer, err := c.spTokenClient.GetBearerToken(ctx, cbbt.JWTOptions{
		ClientID:   c.clientID,
		TenantID:   c.tenantID,
		TimeUTCNow: time.Now().UTC(),
		Duration:   1 * time.Hour,
		NotBefore:  5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("unable to fetch bearer token for SharePoint REST API, error: %w", err)
	}

	site, err := guessSharePointSiteWebURLBase(siteWebURL)
	if err != nil {
		return err
	}

	url, err := url.Parse(site)
	if err != nil {
		return err
	}

	loginName := thingID                // let's say this is just `c:0(.s|true`...
	if strings.Contains(thingID, "@") { // nvm, is an user!
		loginName = strings.Join(append(userLoginName, thingID), "|")
	} else if strings.HasPrefix(thingID, "rolemanager") { // nvm, is a special user like "Everyone except external users"
		loginName = strings.Join(append(rolemanagerLoginName, thingID), "|")
	} else if strings.HasPrefix(thingID, "tenant") { // nvm, is a special user like "Global Administrator"
		loginName = strings.Join(append(tenantLoginName, thingID), "|")
	} else if thingID == "windows" { // nvm, is "All Users (Windows)" for sites that act as Microsoft 365 groups (i.e.: Example Store site)
		loginName = strings.Join(append(allUsersWindows, thingID), "|")
	} else if ok, err := regexp.MatchString(`([^-\s]+)-([^-\s]+)-([^-\s]+)-([^-\s]+)-([^-\s]+)`, thingID); err == nil || ok { // nvm, it may be a M365 group!
		if err != nil {
			return err
		}
		loginName = strings.Join(append(groupLoginName, thingID), "|")
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithHeader("Content-Type", "application/json;odata=verbose"),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithBearerToken(bearer),
		uhttp.WithJSONBody(&SharePointAddThingRequest{
			Metadata:  SharePointAddThingMetadata{Type: "SP.User"},
			LoginName: loginName,
		}),
	}

	url.Path = path.Join(url.Path, fmt.Sprintf("_api/web/sitegroups(%d)/users", groupID))

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
