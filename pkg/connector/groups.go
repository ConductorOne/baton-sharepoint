package connector

import (
	"context"
	"fmt"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sharepoint/pkg/client"
)

const (
	// resourceTypeGroup represents the Entra group resource type.
	resourceTypeGroup = "group"
	// resourceTypeUser represents the user resource type.
	resourceTypeUser = "user"
)

type groupBuilder struct {
	client *client.Client
}

func (g *groupBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return groupResourceType
}

func (g *groupBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}
	if bag.Current() != nil {
		bag.Push(pagination.PageState{ResourceTypeID: groupResourceType.Id})
	}

	sites, err := g.client.ListSites(ctx, bag)
	if err != nil {
		return nil, "", nil, fmt.Errorf("unable to list SharePoint groups, error: %w", err)
	}

	var ret []*v2.Resource

	for _, site := range sites {
		groups, err := g.client.ListGroupsForSite(ctx, site.WebUrl)
		if err != nil {
			return nil, "", nil, err
		}

		for _, group := range groups {
			siteID, err := resource.NewResourceID(siteResourceType, site.WebUrl)
			if err != nil {
				return nil, "", nil, err
			}
			g, err := resource.NewGroupResource(group.Title, groupResourceType, group.ODataID, []resource.GroupTraitOption{
				resource.WithGroupProfile(map[string]interface{}{
					"site":     site.DisplayName,
					"site url": site.WebUrl,
					"id":       group.Id,
				}),
			}, resource.WithParentResourceID(siteID))
			if err != nil {
				return nil, "", nil, fmt.Errorf("cannot create resource from SharePoint group, err: %w", err)
			}
			ret = append(ret, g)
		}
	}

	ntp, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return ret, ntp, nil, nil
}

func (g *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	parts := strings.Split(strings.ToLower(resource.DisplayName), " ")
	kind := strings.TrimSuffix(parts[len(parts)-1], "s") // make the kind singular

	opts := []entitlement.EntitlementOption{
		entitlement.WithDisplayName(fmt.Sprintf("Membership to %s", resource.DisplayName)),
	}

	ent := entitlement.NewAssignmentEntitlement(resource, kind, opts...)

	return []*v2.Entitlement{ent}, "", nil, nil
}

func (g *groupBuilder) Grants(ctx context.Context, rsc *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	securityPrincipals, err := g.client.ListSecurityPrincipalsInGroupByGroupID(ctx, rsc.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	parts := strings.Split(strings.ToLower(rsc.DisplayName), " ")
	kind := strings.TrimSuffix(parts[len(parts)-1], "s")
	var ret []*v2.Grant
	for _, securityPrincipal := range securityPrincipals {
		granted, isGrantable, err := grantHelper(ctx, securityPrincipal, kind, rsc)
		if err != nil {
			return nil, "", nil, fmt.Errorf("groupBuilder.Grants: failed to grant entitlement, error: %w", err)
		}
		if !isGrantable {
			continue
		}
		ret = append(ret, granted)
	}

	return ret, "", nil, nil
}

func grantHelper(ctx context.Context, securityPrincipal client.SecurityPrincipal, kind string, rsc *v2.Resource) (*v2.Grant, bool, error) {
	// Filter out built ins

	if securityPrincipal.LoginName == "SHAREPOINT\\system" {
		return nil, false, nil
	}
	if strings.Contains(securityPrincipal.LoginName, "|rolemanager|") {
		return nil, false, nil
	}
	// Baton ID
	var resourceType string
	var principalName string
	var keyName string

	switch {
	case strings.Contains(securityPrincipal.LoginName, "federateddirectoryclaimprovider"):
		resourceType = resourceTypeGroup // the type of resource on Entra
		parts := strings.Split(securityPrincipal.LoginName, "|")
		if len(parts) < 3 {
			return nil, false, fmt.Errorf("cannot identify group by its ID, error: malformed login name '%s'", securityPrincipal.LoginName)
		}

		principalName = strings.TrimSuffix(parts[2], "_o") // remove suffix in Entra's group ID. Used by SharePoint to indicate "Owners"
	case strings.Contains(securityPrincipal.LoginName, "|tenant|"):
		resourceType = resourceTypeGroup
		parts := strings.Split(securityPrincipal.LoginName, "|")
		if len(parts) < 3 {
			return nil, false, fmt.Errorf("cannot identify group by its ID, error: malformed login name '%s'", securityPrincipal.LoginName)
		}
		principalName = parts[2]
		keyName = "loginName"
	default:
		resourceType = resourceTypeUser
		principalName = securityPrincipal.UserPrincipalName
		keyName = "userPrincipalName"
	}

	principal := &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     principalName,
	}

	if resourceType == resourceTypeGroup {
		return grant.NewGrant(rsc, kind, principal, grant.WithAnnotation(&v2.ExternalResourceMatchID{
			Id: principalName,
		})), true, nil
	} else {
		return grant.NewGrant(rsc, kind, principal, grant.WithAnnotation(&v2.ExternalResourceMatch{
			Key:          keyName,
			Value:        principalName,
			ResourceType: v2.ResourceType_TRAIT_USER,
		})), true, nil
	}
}

func newGroupBuilder(c *client.Client) *groupBuilder {
	return &groupBuilder{client: c}
}
