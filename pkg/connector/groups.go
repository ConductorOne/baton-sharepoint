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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type groupBuilder struct {
	client           *client.Client
	externalSyncMode bool
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
	users, err := g.client.ListUsersInGroupByGroupID(ctx, rsc.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}
	l := ctxzap.Extract(ctx)

	parts := strings.Split(strings.ToLower(rsc.DisplayName), " ")
	kind := strings.TrimSuffix(parts[len(parts)-1], "s")

	var ret []*v2.Grant
	for _, user := range users {
		granted, err := grantHelper(user, g.externalSyncMode, kind, rsc)
		if err != nil {
			if strings.Contains(err.Error(), "unrecognized user '") {
				l.Info("skipping unrecognized principal due to error", zap.Error(err))
				continue
			}
			return nil, "", nil, fmt.Errorf("groupBuilder.Grants: failed to grant entitlement, error: %w", err)
		}
		ret = append(ret, granted)
	}

	return ret, "", nil, nil
}

func grantHelper(user client.SharePointUser, externalSyncMode bool, kind string, rsc *v2.Resource) (*v2.Grant, error) {
	if (user.PrincipalType == client.User && strings.Contains(user.LoginName, "membership") && externalSyncMode) || // If Entra user
		(user.PrincipalType == client.SecurityGroup && strings.Contains(user.LoginName, "federateddirectoryclaimprovider") && externalSyncMode) { // or Entra group
		// Baton ID
		var resourceType string
		var principalName string
		var keyName string

		if strings.Contains(user.LoginName, "federateddirectoryclaimprovider") {
			resourceType = "group" // the type of resource on Entra
			parts := strings.Split(user.LoginName, "|")
			if len(parts) < 3 {
				return nil, fmt.Errorf("cannot identify group by its ID, error: malformed login name '%s'", user.LoginName)
			}

			principalName = strings.TrimSuffix(parts[2], "_o") // remove suffix in Entra's group ID. Used by SharePoint to indicate "Owners"
		} else {
			resourceType = "user"
			principalName = user.UserPrincipalName
			keyName = "userPrincipalName"
		}

		principal := &v2.ResourceId{
			ResourceType: resourceType,
			Resource:     principalName,
		}

		if resourceType == "group" {
			return grant.NewGrant(rsc, kind, principal, grant.WithAnnotation(&v2.ExternalResourceMatchID{
				Id: principalName,
			})), nil
		} else {
			return grant.NewGrant(rsc, kind, principal, grant.WithAnnotation(&v2.ExternalResourceMatch{
				Key:          keyName,
				Value:        principalName,
				ResourceType: v2.ResourceType_TRAIT_USER,
			})), nil
		}
	} else if user.PrincipalType == client.SecurityGroup && !strings.Contains(user.LoginName, "federateddirectoryclaimprovider") { // Regular grants
		id := getReasonableIDfromLoginName(user.LoginName)
		userID, err := resource.NewResourceID(userResourceType, id)
		if err != nil {
			return nil, err
		}

		return grant.NewGrant(rsc, kind, userID), nil
	}

	return nil, fmt.Errorf("grantHelper: unrecognized user '%s' of principal type %s", user.LoginName, user.PrincipalType)
}

func newGroupBuilder(c *client.Client, externalSyncMode bool) *groupBuilder {
	return &groupBuilder{client: c, externalSyncMode: externalSyncMode}
}
