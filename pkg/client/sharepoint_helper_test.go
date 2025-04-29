package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuessSharePointSiteWebURLBaseURLAPI(t *testing.T) {
	input := "https://conductoroneinsulatorone.sharepoint.com/sites/ExampleCrisis/_api/Web/SiteGroups/GetById(5)"
	expected := "https://conductoroneinsulatorone.sharepoint.com/sites/ExampleCrisis"

	result, err := guessSharePointSiteWebURLBase(input)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, result)
}

func TestGuessSharePointSiteWebURLBaseURLContent(t *testing.T) {
	input := "https://conductoroneinsulatorone.sharepoint.com/sites/ExampleStore/SitePages/Forms/ByAuthor.aspx"
	expected := "https://conductoroneinsulatorone.sharepoint.com/sites/ExampleStore"

	result, err := guessSharePointSiteWebURLBase(input)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, result)
}
