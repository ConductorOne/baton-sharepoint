![Baton Logo](./baton-logo.png)

# `baton-sharepoint` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-sharepoint.svg)](https://pkg.go.dev/github.com/conductorone/baton-sharepoint) ![main ci](https://github.com/conductorone/baton-sharepoint/actions/workflows/main.yaml/badge.svg)

`baton-sharepoint` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-sharepoint
baton-sharepoint
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-sharepoint:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-sharepoint/cmd/baton-sharepoint@main

baton-sharepoint

baton resources
```

# Data Model

`baton-sharepoint` will pull down information about the following resources:
- Users
- Groups
- Sites

# Permissions

- SharePoint
  - `Sites.FullControl.All` (Application): Have full control of all site collections
  - `Sites.Read.All` (Application): Read items in all site collections
  - `User.Read.All` (Application): Read user profiles
- Microsoft Graph
  - `Sites.Read.All` (Application): Read items in all site collections

## SharePoint requirements

Please make a self-signed certificate and upload it to your registered
application at *Certificates & secrets* > *Certificates*. Under
GNU/Linux you can make a certificate with the script
`./scripts/generate-self-signed-certificate.sh`.

To generate a self-signed certificate under Microsoft Windows, use
this
[script](https://github.com/LucasMarangon/Azure_Oauth_JWT/blob/a66a55737eeae775c0bbe19dfbfc04e292fc7702/Create-SelfSignedCertificate.ps1)
(update the variables in there accordingly).

> [!WARNING]
> Please note that a third-party maintains that Powershell script

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-sharepoint` Command Line Usage

```
baton-sharepoint

Usage:
  baton-sharepoint [flags]
  baton-sharepoint [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  config             Get the connector config schema
  help               Help about any command

Flags:
      --azure-client-id string                           required: Azure Client ID ($BATON_AZURE_CLIENT_ID)
      --azure-client-secret string                       required: Azure Client Secret ($BATON_AZURE_CLIENT_SECRET)
      --azure-graph-domain string                        Domain for Microsoft Graph API ($BATON_AZURE_GRAPH_DOMAIN) (default "graph.microsoft.com")
      --azure-tenant-id string                           required: Azure Tenant ID ($BATON_AZURE_TENANT_ID)
      --client-id string                                 The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string                             The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --external-resource-c1z string                     The path to the c1z file to sync external baton resources with ($BATON_EXTERNAL_RESOURCE_C1Z)
      --external-resource-entitlement-id-filter string   The entitlement that external users, groups must have access to sync external baton resources ($BATON_EXTERNAL_RESOURCE_ENTITLEMENT_ID_FILTER)
  -f, --file string                                      The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                                             help for baton-sharepoint
      --log-format string                                The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string                                 The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --otel-collector-endpoint string                   The endpoint of the OpenTelemetry collector to send observability data to (used for both tracing and logging if specific endpoints are not provided) ($BATON_OTEL_COLLECTOR_ENDPOINT)
      --pfx-certificate string                           required: Base64-encoded PFX certificate ($BATON_PFX_CERTIFICATE)
      --pfx-certificate-password string                  required: Password of the PFX certificate ($BATON_PFX_CERTIFICATE_PASSWORD)
  -p, --provisioning                                     This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --sharepoint-domain string                         required: Domain of SharePoint ($BATON_SHAREPOINT_DOMAIN)
      --skip-full-sync                                   This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --ticketing                                        This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                                          version for baton-sharepoint

Use "baton-sharepoint [command] --help" for more information about a command.
```
