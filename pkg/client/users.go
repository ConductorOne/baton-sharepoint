package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetUserPrincipalNameFromUserID fetch the user principal name of a user by its ID.
func (c *Client) GetUserPrincipalNameFromUserID(ctx context.Context, userID string) (string, error) {
	defaultValues := url.Values{}
	defaultValues.Set("$select", strings.Join([]string{"id", "userPrincipalName"}, ","))

	targetURL := c.buildURL("users/"+userID, defaultValues)

	var resp UserPrincipalNameResponse
	err := c.query(ctx, makeGraphReadScopes(c.GraphDomain), http.MethodGet, targetURL, nil, &resp)
	if err != nil {
		return "", fmt.Errorf("GetUserPrincipalName: request failed, error: %w", err)
	}

	return resp.UserPrincipalName, nil
}
