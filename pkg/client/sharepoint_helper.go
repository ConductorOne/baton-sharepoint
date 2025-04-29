package client

import (
	"net/url"
	"strings"
)

// guessSharePointSiteWebURLBase gives you the web URL of a SharePoint site
//
// Imagine you have the following URL:
//
//	https://conductoroneinsulatorone.sharepoint.com/sites/ExampleStore/SitePages/Forms/ByAuthor.aspx
//	https://conductoroneinsulatorone.sharepoint.com/sites/ExampleCrisis/_api/Web/SiteGroups/GetById(5)
//
// This function would give you:
//
//	https://conductoroneinsulatorone.sharepoint.com/sites/ExampleStore/
//	https://conductoroneinsulatorone.sharepoint.com/sites/ExampleCrisis/
//
// It helps to figure out what's the root of a site and then do something else with
// that string, for example, build the URL path for an API call, etc.
func guessSharePointSiteWebURLBase(site string) (string, error) {
	web, err := url.Parse(site)
	if err != nil {
		return "", err
	}

	parts := strings.Split(web.Path, "/")
	for index := len(parts) - 1; index > 0; index-- {
		part := parts[index]
		if strings.HasPrefix(part, "_") { // for `_api` or `_layout`
			web.Path = strings.Join(parts[:index], "/")
			break
		}
		if part == "sites" { // TODO(shackra): figure out how to handle sub-sites
			web.Path = strings.Join(parts[:index+2], "/")
			break
		}
	}

	return web.String(), nil
}
