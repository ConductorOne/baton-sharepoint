package connector

import "strings"

func getReasonableIDfromLoginName(loginName string) string {
	userID := loginName

	parts := strings.SplitN(loginName, "|", 2)
	if len(parts) == 2 && parts[len(parts)-1] != "true" { // `c:0(.s|true` is the ID for `Everyone`
		userID = parts[1]
	}

	return userID
}
