package sde

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportBlueprints_Smoketest(t *testing.T) {
	blueprintsYAMLFile, err := os.Open("../../test/testdata/blueprints.yaml")
	if err != nil {
		wd, _ := os.Getwd()
		t.Fatalf("CWD is `%v` Failed to open blueprint YAML test data: %v", wd, err)
	}

	contents, err := ioutil.ReadAll(blueprintsYAMLFile)
	if err != nil {
		t.Fatalf("Failed to read blueprint YAML test data: %v", err)
	}
	t.Logf("loaded %d bytes from blueprint yaml", len(contents))

	blueprints, err := ImportBlueprints([]byte(contents))

	var totalBlueprints int
	for _, v := range blueprints {
		totalBlueprints = totalBlueprints + len(v)
	}
	t.Logf("Loaded %d blueprints", totalBlueprints)

	assert.Nil(t, err)

	for _, productBlueprints := range blueprints {
		for _, blueprint := range productBlueprints {
			t.Log(blueprint.BlueprintTypeID)
		}
	}
}
