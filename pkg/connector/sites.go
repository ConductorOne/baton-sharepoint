package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sharepoint/pkg/client"
)

type siteBuilder struct {
	client *client.Client
}

func (o *siteBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return siteResourceType
}

func (o *siteBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}

	if bag.Current() != nil {
		bag.Push(pagination.PageState{ResourceTypeID: siteResourceType.Id})
	}

	sites, err := o.client.ListSites(ctx, bag)
	if err != nil {
		return nil, "", nil, fmt.Errorf("listBuilder.List: cannot list Sites, error: %w", err)
	}

	npt, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	ret, err := convertSites2Resources(sites)
	if err != nil {
		return nil, "", nil, fmt.Errorf("listBuilder.List: cannot convert Sites to resources, error: %w", err)
	}

	return ret, npt, nil, nil
}

func (o *siteBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	opts := []entitlement.EntitlementOption{
		entitlement.WithDisplayName(fmt.Sprintf("Membership to %s", resource.DisplayName)),
	}

	ent := entitlement.NewAssignmentEntitlement(resource, "member", opts...)

	return []*v2.Entitlement{ent}, "", nil, nil
}

func (o *siteBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	// TODO(shackra): implement this
	return nil, "", nil, nil
}

func newListBuilder(c *client.Client) *siteBuilder {
	return &siteBuilder{client: c}
}

func convertSite2Resource(site client.Site) (*v2.Resource, error) {
	profile := map[string]any{
		"display name":  site.DisplayName,
		"name":          site.Name,
		"url":           site.WebUrl,
		"personal site": site.IsPersonalSite,
	}

	opts := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}

	rsc, err := resource.NewGroupResource(site.DisplayName, siteResourceType, site.ID, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot make resource from Site '%s', error: %w", site.DisplayName, err)
	}

	return rsc, nil
}

func convertSites2Resources(sites []client.Site) ([]*v2.Resource, error) {
	var ret []*v2.Resource

	for _, site := range sites {
		rsc, err := convertSite2Resource(site)
		if err != nil {
			return nil, err
		}

		ret = append(ret, rsc)
	}

	return ret, nil
}
