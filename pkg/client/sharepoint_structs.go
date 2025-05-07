package client

// UserOrGroupPrincipalType Specifies the type of a principal for either Users or Groups
// documentation: https://learn.microsoft.com/en-us/previous-versions/office/sharepoint-csom/ee541430(v=office.15)#members
// Note(shackra): Please note that the bitwise operation is not implemented yet.
type UserOrGroupPrincipalType int

const (
	// Enumeration whose value specifies no principal type.
	None UserOrGroupPrincipalType = iota
	// Enumeration whose value specifies a user as the principal type.
	User
	// Enumeration whose value specifies a distribution list as the principal type.
	DistributionList
	// Enumeration whose value specifies a security group as the principal type.
	SecurityGroup = 4
	// Enumeration whose value specifies a group (2) as the principal type.
	SharePointGroup = 8
	// Enumeration whose value specifies all principal types.
	All = 15
)

func (t UserOrGroupPrincipalType) String() string {
	value := ""
	switch t {
	case None:
		value = "None"
	case User:
		value = "User"
	case DistributionList:
		value = "Distribution List"
	case SecurityGroup:
		value = "Security Group"
	case SharePointGroup:
		value = "SharePoint Group"
	case All:
		value = "All"
	}

	return value
}

// SharePointSiteGroup is a SP.Group
// documentation: https://learn.microsoft.com/en-us/previous-versions/office/developer/sharepoint-rest-reference/dn531432(v=office.15)#group-properties
type SharePointSiteGroup struct {
	ODataID   string `json:"odata.id"`
	ODataType string `json:"odata.type"`
	// Gets a value that specifies the member identifier for the user or group.
	Id int `json:"Id"`
	// Implement this abstract property to get a string that contains the login name of the user.
	LoginName string `json:"LoginName"`
	// Gets or sets a value that specifies the name of the principal.
	Title string `json:"Title"`
	// Gets the name for the owner of this group.
	OwnerTitle string `json:"OwnerTitle"`
	// Gets or sets a value that indicates whether the group members can edit membership in the group.
	AllowMembersEditMembership bool `json:"AllowMembersEditMembership"`
	// Gets or sets a Boolean value that specifies whether to allow users to request membership in the group and to allow users to request to leave the group.
	AllowRequestToJoinLeave bool `json:"AllowRequestToJoinLeave"`
	// Gets or sets a Boolean value that specifies whether users are automatically added or removed when they make a request.
	AutoAcceptRequestToJoinLeave bool `json:"AutoAcceptRequestToJoinLeave"`
	// Gets or sets the description for the group.
	Description *string `json:"Description"`
	// Gets or sets a Boolean value that specifies whether only group members are allowed to view the list of members in the group.
	OnlyAllowMembersViewMembership bool `json:"OnlyAllowMembersViewMembership"`
	// Gets or sets the e-mail address to which membership requests are sent.
	RequestToJoinLeaveEmailSetting string `json:"RequestToJoinLeaveEmailSetting"`
	// Gets a value containing the type of the principal.
	PrincipalType int `json:"PrincipalType"`
}

// SharePointUserId is a SP.UserIdInfo
// documentation: (check the link to the documentation on the `SharePointUser` struct)
type SharePointUserId struct {
	NameId       string
	NameIdIssuer string
}

// SharePointSiteUser is a SP.User
// documentation: https://learn.microsoft.com/en-us/previous-versions/office/developer/sharepoint-rest-reference/dn531432(v=office.15)#user-properties
type SharePointUser struct {
	ODataType     string                   `json:"odata.type"`
	ODataID       string                   `json:"odata.id"`
	Id            int                      `json:"Id"`            // Gets a value that specifies the member identifier for the user or group.
	Title         string                   `json:"Title"`         // Gets or sets a value that specifies the name of the principal.
	Email         string                   `json:"Email"`         // Gets or sets the email address of the user.
	IsSiteAdmin   bool                     `json:"IsSiteAdmin"`   // Gets or sets a Boolean value that specifies whether the user is a site collection administrator.
	UserId        *SharePointUserId        `json:"UserId"`        // Gets the information of the user that contains the user's name identifier and the issuer of the user's name identifier.
	LoginName     string                   `json:"LoginName"`     // Gets the login name of the user.
	IsHiddenInUI  bool                     `json:"IsHiddenInUI"`  // Gets a value that indicates whether this member should be hidden in the UI.
	PrincipalType UserOrGroupPrincipalType `json:"PrincipalType"` // Gets a value containing the type of the principal. Represents a bitwise SP.PrincipalType

	// Set if the SharePoint site reports them as such
	IsEmailAuthenticationGuestUser bool   `json:"IsEmailAuthenticationGuestUser"`
	IsShareByEmailGuestUser        bool   `json:"IsShareByEmailGuestUser"`
	UserPrincipalName              string `json:"UserPrincipalName"`
}

type SharePointAddThingMetadata struct {
	Type string `json:"type"`
}

type SharePointAddThingRequest struct {
	Metadata  SharePointAddThingMetadata `json:"__metadata"`
	LoginName string                     `json:"LoginName"`
}

type SharePointEnsureThingRequest struct {
	LogonName string `json:"logonName"`
}

// Local Variables:
// go-tag-args: ("-transform" "pascalcase")
// End:
