package connector

import (
	"context"
	"fmt"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sharepoint/pkg/client"
)

type Connector struct {
	client *client.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newSiteBuilder(d.client),
		newGroupBuilder(d.client),
		newSecurityPrincipalBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Baton SharePoint",
		Description: "Baton connector for Microsoft® SharePoint™",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
// cert parameter should contain the raw content of a PFX certificate file.
func New(ctx context.Context, tenantID, clientID, clientSecret, graphDomain, sharepointDomain, cert string,
	certpassword string, syncSharePointHomeOrgLinks bool,
) (*Connector, error) {
	c, err := client.New(ctx, tenantID, clientID, clientSecret, graphDomain, sharepointDomain, cert, certpassword, syncSharePointHomeOrgLinks)
	if err != nil {
		return nil, fmt.Errorf("failed to make connector, error: %w", err)
	}

	return &Connector{client: c}, nil
}
