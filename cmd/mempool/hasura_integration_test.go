package main

import (
	"github.com/dipdup-net/go-lib/hasura"
	"testing"
)

func TestIntegration_HasuraMetadata(t *testing.T) {
	// read config
	configPath := "../../build/dipdup.testnet.yml"              // todo: Fix paths
	expectedMetadataPath := "../../build/expected_metadata.yml" // todo: Fix paths

	hasura.TestExpectedMetadataWithActual(t, configPath, expectedMetadataPath)
}
