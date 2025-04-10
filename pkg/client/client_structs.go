package client

type GetAllSitesResponse struct {
	Value    []Site `json:"value"`
	NextLink string `json:"@odata.nextLink"`
}

type GetUserInformationListItemsResponse struct {
	NextLink string `json:"@odata.nextLink"`
	// TODO(shackra): try against a real SharePoint instance to figure the structure of the response
}

// Local Variables:
// go-tag-args: ("-transform" "camelcase")
// End:
