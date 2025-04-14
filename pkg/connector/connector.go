package connector

import (
	"context"
	"fmt"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sharepoint/pkg/client"
	cert_based_bearer_token "github.com/conductorone/baton-sharepoint/pkg/client/cert-based-bearer-token"
)

type Connector struct {
	client *client.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newListBuilder(d.client),
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
		Description: "",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, tenantID, clientID, clientSecret, graphDomain, sharepointDomain, cert, certpassword string) (*Connector, error) {
	spc, err := cert_based_bearer_token.New(ctx, sharepointDomain, cert, certpassword)
	if err != nil {
		return nil, fmt.Errorf("failed to make connector, error: %w", err)
	}

	c, err := client.New(ctx, spc, tenantID, clientID, clientSecret, graphDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to make connector, error: %w", err)
	}

	return &Connector{client: c}, nil
}
