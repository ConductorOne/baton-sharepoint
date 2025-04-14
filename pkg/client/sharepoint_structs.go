package client

type SharePointSiteGroup struct {
	ODataID                        string  `json:"odata.id"`
	ODataType                      string  `json:"odata.type"`
	Id                             int     `json:"Id"`
	LoginName                      string  `json:"LoginName"`
	Title                          string  `json:"Title"`
	OwnerTitle                     string  `json:"OwnerTitle"`
	AllowMembersEditMembership     bool    `json:"AllowMembersEditMembership"`
	AllowRequestToJoinLeave        bool    `json:"AllowRequestToJoinLeave"`
	AutoAcceptRequestToJoinLeave   bool    `json:"AutoAcceptRequestToJoinLeave"`
	Description                    *string `json:"Description"`
	OnlyAllowMembersViewMembership bool    `json:"OnlyAllowMembersViewMembership"`
	RequestToJoinLeaveEmailSetting string  `json:"RequestToJoinLeaveEmailSetting"`
}

type SharePointSiteUserId struct {
	NameId       string
	NameIdIssuer string
}

type SharePointSiteUser struct {
	ODataType                      string                `json:"odata.type"`
	ODataID                        string                `json:"odata.id"`
	Id                             int                   `json:"Id"`
	LoginName                      string                `json:"LoginName"`
	Title                          string                `json:"Title"`
	Email                          string                `json:"Email"`
	IsEmailAuthenticationGuestUser bool                  `json:"IsEmailAuthenticationGuestUser"`
	IsShareByEmailGuestUser        bool                  `json:"IsShareByEmailGuestUser"`
	IsSiteAdmin                    bool                  `json:"IsSiteAdmin"`
	UserId                         *SharePointSiteUserId `json:"UserId"`
	UserPrincipalName              string                `json:"UserPrincipalName"`
}

// Local Variables:
// go-tag-args: ("-transform" "pascalcase")
// End:
