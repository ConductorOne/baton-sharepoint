package errorexplained

import (
	"fmt"
	"strings"
)

type ErrorExplained struct {
	ErrorType   string `json:"error"`
	Description string `json:"error_description"`
	Codes       []int  `json:"error_codes"`
	URI         string `json:"error_uri"`
}

func (t *ErrorExplained) Message() string {
	if strings.Contains(t.Description, "Reason - The key was not found., Thumbprint of key used by client") {
		return t.ErrorType + ": certificate used by client is unknown to the server, did you uploaded the CRT certificate at 'App Registration'?"
	}
	if strings.Contains(t.Description, "AADSTS900023: Specified tenant identifier") {
		return t.ErrorType + ": the 'Directory (Tenant) ID' specified is invalid"
	}
	if strings.Contains(t.Description, "AADSTS7000215: Invalid client secret provided") {
		return t.ErrorType + ": the 'Client Secret' specified is invalid. Please ensure *you did not* pass the client secret's ID instead!"
	}

	return t.ErrorType + ": " + t.Description
}

func WhatErrorToReturn(expl ErrorExplained, err error) error {
	if expl.Description == "" {
		return err
	}

	return fmt.Errorf("Entra/SharePoint API error: %s", expl.Message())
}

// Local Variables:
// go-tag-args: ("-transform" "snakecase")
// End:
