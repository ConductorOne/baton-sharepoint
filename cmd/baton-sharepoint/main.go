package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/conductorone/baton-sharepoint/pkg/connector"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-sharepoint",
		getConnector,
		field.Configuration{
			Fields:      ConfigurationFields,
			Constraints: FieldRelationships,
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)
	if err := ValidateConfig(v); err != nil {
		return nil, err
	}

	var certContent string
	certFilePath := v.GetString(CertFilePathField.FieldName)

	// If certificate file path is provided, read the certificate content from the file
	if certFilePath != "" {
		certBytes, err := os.ReadFile(certFilePath)
		if err != nil {
			l.Error("error reading certificate file", zap.Error(err), zap.String("certFilePath", certFilePath))
			return nil, fmt.Errorf("failed to read certificate file: %w", err)
		}
		certContent = string(certBytes)
	} else {
		// Otherwise use the certificate content provided directly
		certContent = v.GetString(CertPfxField.FieldName)
	}

	cb, err := connector.New(
		ctx,
		v.GetString(TenantIDField.FieldName),
		v.GetString(ClientIDField.FieldName),
		v.GetString(ClientSecretField.FieldName),
		v.GetString(GraphDomainField.FieldName),
		v.GetString(SharePointDomainField.FieldName),
		certContent,
		v.GetString(CertPasswordField.FieldName),
		v.GetString("external-resource-c1z") != "", // NOTE(shackra): expect problems if that flag is renamed
		v.GetBool(SyncOrgLinkGroupsField.FieldName),
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return connector, nil
}
