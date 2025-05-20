package connector

import (
	"context"
	"fmt"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sharepoint/pkg/client"
)

type securityPrincipalBuilder struct {
	client *client.Client
}

func (s *securityPrincipalBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return securityPrincipalResourceType
}

func (s *securityPrincipalBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}
	if bag.Current() != nil {
		bag.Push(pagination.PageState{ResourceTypeID: groupResourceType.Id})
	}

	sites, err := s.client.ListSites(ctx, bag)
	if err != nil {
		return nil, "", nil, fmt.Errorf("unable to list SharePoint security principals, error: %w", err)
	}

	var ret []*v2.Resource

	for _, site := range sites {
		users, err := s.client.ListSecurityPrincipals(ctx, site.WebUrl)
		if err != nil {
			return nil, "", nil, err
		}

		for _, user := range users {
			// ignore Entra users, Microsoft 365 Groups, Entra groups and "system" users
			if user.PrincipalType == client.SecurityGroup &&
				!strings.Contains(user.LoginName, "federateddirectoryclaimprovider") &&
				!strings.Contains(user.LoginName, "|tenant|") {
				spResource, err := resource.NewGroupResource(
					user.Title,
					securityPrincipalResourceType,
					user.LoginName,
					nil,
				)
				if err != nil {
					return nil, "", nil, fmt.Errorf("cannot create resource from SharePoint security principal, err: %w", err)
				}
				ret = append(ret, spResource)
			}
		}
	}

	ntp, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return ret, ntp, nil, nil
}

func (s *securityPrincipalBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (s *securityPrincipalBuilder) Grants(ctx context.Context, rsc *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newSecurityPrincipalBuilder(c *client.Client) *securityPrincipalBuilder {
	return &securityPrincipalBuilder{client: c}
}
