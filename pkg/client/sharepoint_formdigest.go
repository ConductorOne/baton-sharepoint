package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	cbbt "github.com/conductorone/baton-sharepoint/pkg/client/cert-based-bearer-token"
)

const (
	formDigestDateLayout = "02 Jan 2006 15:04:05 -0700"
)

type formDigestResponse struct {
	FormDigestTimeoutSeconds int    `json:"FormDigestTimeoutSeconds"`
	FormDigestValue          string `json:"FormDigestValue"`
	LibraryVersion           string `json:"LibraryVersion"`
	SiteFullURL              string `json:"SiteFullUrl"`
	WebFullURL               string `json:"WebFullUrl"`
}

func (c *Client) getFormDigestValue(ctx context.Context, site string) (string, error) {
	c.digestMutex.Lock()
	defer c.digestMutex.Unlock()

	expired := c.digestValueExpiration[site]
	now := time.Now().UTC()
	if !expired.IsZero() && now.Before(expired) {
		return c.digestValuePerSite[site], nil
	}

	// obtain the form digest value
	siteUrl, err := url.Parse(site)
	if err != nil {
		return "", err
	}

	bearer, err := c.spTokenClient.GetBearerToken(ctx, cbbt.JWTOptions{
		ClientID:   c.clientID,
		TenantID:   c.tenantID,
		TimeUTCNow: time.Now().UTC(),
		Duration:   1 * time.Hour,
		NotBefore:  5 * time.Minute,
	})
	if err != nil {
		return "", fmt.Errorf("unable to fetch bearer token for SharePoint REST API, error: %w", err)
	}

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithBearerToken(bearer),
		uhttp.WithFormBody(""),
	}

	siteUrl.Path = path.Join(siteUrl.Path, "_api/contextinfo")

	req, err := c.http.NewRequest(ctx, http.MethodPost, siteUrl, reqOpts...)
	if err != nil {
		return "", err
	}

	var data formDigestResponse
	resp, err := c.http.Do(req, uhttp.WithJSONResponse(&data))
	if err != nil {
		return "", err
	}

	resp.Body.Close()

	parts := strings.Split(data.FormDigestValue, ",")
	if len(parts) < 2 {
		return "", fmt.Errorf("malformed digest value received '%s'", data.FormDigestValue)
	}

	digestTime, err := time.Parse(formDigestDateLayout, parts[1])
	if err != nil {
		return "", err
	}
	expiresAt := digestTime.Add(time.Second * time.Duration(data.FormDigestTimeoutSeconds))

	c.digestValueExpiration[data.SiteFullURL] = expiresAt
	c.digestValuePerSite[data.SiteFullURL] = data.FormDigestValue

	return data.FormDigestValue, nil
}

// Local Variables:
// go-tag-args: ("-transform" "pascalcase")
// End:
