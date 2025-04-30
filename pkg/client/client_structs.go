package client

type GetAllSitesResponse struct {
	Value    []Site `json:"value"`
	NextLink string `json:"@odata.nextLink"`
}

type ListGroupsForSiteResponse struct {
	Value []SharePointSiteGroup `json:"value"`
}

type ListUsersInGroupByGroupIDResponse struct {
	Value []SharePointUser `json:"value"`
}

type ListUsersResponse struct {
	Value []SharePointUser `json:"value"`
}

// Local Variables:
// go-tag-args: ("-transform" "camelcase")
// End:
