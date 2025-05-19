package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var siteResourceType = &v2.ResourceType{
	Id:          "site",
	DisplayName: "Site",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}

var groupResourceType = &v2.ResourceType{
	Id:          "sharepoint_group",
	DisplayName: "SharePoint Group",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}

var securityPrincipalResourceType = &v2.ResourceType{
	Id:          "security_principal",
	DisplayName: "Security Principal",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}
