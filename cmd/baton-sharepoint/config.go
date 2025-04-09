package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	TenantIDField     = field.StringField("tenant-id", field.WithDescription("Azure Tenant ID"), field.WithRequired(true))
	ClientIDField     = field.StringField("client-id", field.WithDescription("Azure Client ID"), field.WithRequired(true))
	ClientSecretField = field.StringField("client-secret", field.WithDescription("Azure Client Secret"), field.WithRequired(true))
	GraphDomainField  = field.StringField("graph-domain", field.WithDescription("Domain for Microsoft Graph API"), field.WithDefaultValue("graph.microsoft.com"))
)

var (
	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{
		TenantIDField,
		ClientIDField,
		ClientSecretField,
		GraphDomainField,
	}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	return nil
}
