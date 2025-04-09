package client

// The root facet indicates that an object is the top-most one in its
// hierarchy. The presence (non-null) of the facet value indicates
// that the object is the root. A null (or missing) value indicates
// the object is not the root.
//
// > Note: While this facet is empty today, in future API revisions
// > the facet may be populated with additional properties.
type Root struct {
	ODataType string `json:"odata.type"`
}

type Site struct {
	ID             string         `json:"id"`             // The unique identifier of the item. Read-only.
	Name           string         `json:"name"`           // The name/title of the item.
	DisplayName    string         `json:"displayName"`    // The full title for the site. Read-only.
	IsPersonalSite bool           `json:"isPersonalSite"` // Identifies whether the site is personal or not. Read-only.
	SiteCollection SiteCollection `json:"siteCollection"` // Provides details about the site's site collection. Available only on the root site. Read-only.
	WebUrl         string         `json:"webUrl"`         // URL that displays the item in the browser. Read-only.
	Root           *Root          `json:"root"`           // If present, provides the root site in the site collection. Read-only.
}

type SiteCollection struct {
	Hostname         string `json:"hostname"`         // The hostname for the site
	DataLocationCode string `json:"dataLocationCode"` // collection. Read-only. The geographic region code for where this site collection resides. Only present for multi-geo tenants. Read-only.
	Root             *Root  `json:"root"`             // If present, indicates that this is a root site collection in SharePoint. Read-only.
}

// Local Variables:
// go-tag-args: ("-transform" "camelcase")
// End:
