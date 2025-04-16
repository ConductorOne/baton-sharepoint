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
			siteID, err := resource.NewResourceID(siteResourceType, site.ID)
			if err != nil {
				return nil, "", nil, err
			}
			g, err := resource.NewGroupResource(group.Title, groupResourceType, group.ODataID, []resource.GroupTraitOption{
				resource.WithGroupProfile(map[string]interface{}{"site": site.DisplayName, "site_url": site.WebUrl}),
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
	opts := []entitlement.EntitlementOption{
		entitlement.WithDisplayName(fmt.Sprintf("Membership to %s", resource.DisplayName)),
	}

	ent := entitlement.NewAssignmentEntitlement(resource, "member", opts...)

	return []*v2.Entitlement{ent}, "", nil, nil
}

func (g *groupBuilder) Grants(ctx context.Context, rsc *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	users, err := g.client.ListUsersInGroupByGroupID(ctx, rsc.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Grant
	for _, user := range users {
		if (user.PrincipalType == 1 && strings.Index(user.LoginName, "membership") != -1) || (user.PrincipalType == 4 && strings.Index(user.LoginName, "federateddirectoryclaimprovider") != -1) {
			var resourceType string
			var principalName string

			if strings.Index(user.LoginName, "federateddirectoryclaimprovider") != -1 {
				resourceType = "group"     // TODO(shackra): check this is the ID of that resource in baton-microsoft-entra
				principalName = user.Email // groups don't have UserPrincipalName field set
			} else {
				resourceType = "user" // TODO(shackra): check this is the ID of that resource in baton-microsoft-entra
				principalName = user.UserPrincipalName
			}
			principal := &v2.ResourceId{
				ResourceType: resourceType,
				Resource:     principalName,
			}

			ret = append(ret, grant.NewGrant(rsc, "member", principal, grant.WithAnnotation(&v2.ExternalResourceMatch{
				Key:   "email",
				Value: principalName,
			})))
		} else if user.PrincipalType == 4 && strings.Index(user.LoginName, "federateddirectoryclaimprovider") == -1 {
			userID, err := resource.NewResourceID(userResourceType, user.ODataID)
			if err != nil {
				return nil, "", nil, err
			}
			ret = append(ret, grant.NewGrant(rsc, "member", userID))
		}
	}

	return ret, "", nil, nil
}

func newGroupBuilder(c *client.Client) *groupBuilder {
	return &groupBuilder{client: c}
}
