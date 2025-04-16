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
	PrincipalType                  int     `json:"PrincipalType"`
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
	PrincipalType                  int                   `json:"PrincipalType"`
}

type SharePointUser struct {
	ODataType     string                `json:"odata.type"`
	ODataID       string                `json:"odata.id"`
	Id            int                   `json:"Id"`            // Gets a value that specifies the member identifier for the user or group.
	Title         string                `json:"Title"`         // Gets or sets a value that specifies the name of the principal.
	Email         string                `json:"Email"`         // Gets or sets the email address of the user.
	IsSiteAdmin   bool                  `json:"IsSiteAdmin"`   // Gets or sets a Boolean value that specifies whether the user is a site collection administrator.
	UserId        *SharePointSiteUserId `json:"UserId"`        // Gets the information of the user that contains the user's name identifier and the issuer of the user's name identifier.
	PrincipalType int                   `json:"PrincipalType"` // Gets a value containing the type of the principal. Represents a bitwise SP.PrincipalType value: None = 0; User = 1; DistributionList = 2; SecurityGroup = 4; SharePointGroup = 8; All = 15.
	LoginName     string                `json:"LoginName"`     // Gets the login name of the user.
	IsHiddenInUI  bool                  `json:"IsHiddenInUI"`  // Gets a value that indicates whether this member should be hidden in the UI.
}

// Local Variables:
// go-tag-args: ("-transform" "pascalcase")
// End:
