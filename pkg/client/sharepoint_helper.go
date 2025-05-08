package client

import (
	"net/url"
	"regexp"
	"strings"
)

// GuessSharePointSiteWebURLBase gives you the web URL of a SharePoint site
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
func GuessSharePointSiteWebURLBase(site string) (string, error) {
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

var (
	userLoginName        = []string{"i:0#.f", "membership"}
	groupLoginName       = []string{"c:0o.c", "federateddirectoryclaimprovider"}
	rolemanagerLoginName = []string{"c:0-.f"}
	tenantLoginName      = []string{"c:0t.c"}
	allUsersWindows      = []string{"c:0!.s"}

	looksLikeUUID = regexp.MustCompile(`([^-\s]+)-([^-\s]+)-([^-\s]+)-([^-\s]+)-([^-\s]+)`)
)

func guessFullLoginName(partialLoginName string) string {
	loginName := partialLoginName                // let's say this is just `c:0(.s|true`...
	if strings.Contains(partialLoginName, "@") { // nvm, is an user!
		loginName = strings.Join(append(userLoginName, partialLoginName), "|")
	} else if strings.HasPrefix(partialLoginName, "rolemanager") { // nvm, is a special user like "Everyone except external users"
		loginName = strings.Join(append(rolemanagerLoginName, partialLoginName), "|")
	} else if strings.HasPrefix(partialLoginName, "tenant") { // nvm, is a Microsoft 365 group!
		loginName = strings.Join(append(tenantLoginName, partialLoginName), "|")
	} else if partialLoginName == "windows" { // nvm, is "All Users (Windows)" for sites that act as Microsoft 365 groups (i.e.: Example Store site)
		loginName = strings.Join(append(allUsersWindows, partialLoginName), "|")
	} else if looksLikeUUID.MatchString(partialLoginName) { // nvm, it may be a M365 group!
		loginName = strings.Join(append(groupLoginName, partialLoginName), "|")
	}

	return loginName
}
