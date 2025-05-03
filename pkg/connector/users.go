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

type userBuilder struct {
	client *client.Client
}

func (u *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}
	if bag.Current() != nil {
		bag.Push(pagination.PageState{ResourceTypeID: groupResourceType.Id})
	}

	sites, err := u.client.ListSites(ctx, bag)
	if err != nil {
		return nil, "", nil, fmt.Errorf("unable to list SharePoint users, error: %w", err)
	}

	var ret []*v2.Resource

	for _, site := range sites {
		users, err := u.client.ListSharePointUsers(ctx, site.WebUrl)
		if err != nil {
			return nil, "", nil, err
		}

		for _, user := range users {
			// ignore Entra users, Microsoft 365 Groups and "system" users
			if user.PrincipalType == client.SecurityGroup && !strings.Contains(user.LoginName, "federateddirectoryclaimprovider") {
				userID := getReasonableIDfromLoginName(user.LoginName)
				ur, err := resource.NewUserResource(user.Title, userResourceType, userID, []resource.UserTraitOption{
					resource.WithUserProfile(map[string]interface{}{
						"site":            site.DisplayName,
						"site url":        site.WebUrl,
						"login name":      user.LoginName,
						"site id":         user.Id,
						"is site admin":   user.IsSiteAdmin,
						"is hidden in ui": user.IsHiddenInUI,
					}),
					resource.WithEmail(user.Email, true),
				})
				if err != nil {
					return nil, "", nil, fmt.Errorf("cannot create resource from SharePoint user, err: %w", err)
				}
				ret = append(ret, ur)
			}
		}
	}

	ntp, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return ret, ntp, nil, nil
}

func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (g *userBuilder) Grants(ctx context.Context, rsc *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(c *client.Client) *userBuilder {
	return &userBuilder{client: c}
}
