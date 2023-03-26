package main

import (
	"context"
	libCfg "github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/hasura"
	"github.com/dipdup-net/mempool/cmd/mempool/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"testing"
)

func TestIntegration_HasuraMetadata(t *testing.T) {
	// read config
	configPath := "../../build/dipdup.testnet.yml" // todo: Fix paths
	var cfg config.Config
	if err := libCfg.Parse(configPath, &cfg); err != nil {
		log.Err(err).Msg("") // or fail
		return
	}

	integrationHelpers := hasura.NewIntegrationHelpers(cfg.Hasura)

	ctx, _ := context.WithCancel(context.Background())

	metadata, err := integrationHelpers.GetMetadata(ctx)
	if err != nil {
		t.Fatalf("Error with getting hasura metadata %e", err)
	}

	expectedMetadataDefinition := "../../build/expected_metadata.yml" // todo: Fix paths
	expectedMetadata, err := integrationHelpers.ParseExpectedMetadata(expectedMetadataDefinition)
	if err != nil {
		t.Fatalf("Error with parsing expected metadata: %e", err)
	}

	// Go through `expectedMetadata` and assert that each object
	// in that array is in `metadata` with corresponding columns.
	for _, expectedTable := range expectedMetadata.Tables {
		metadataTable, err := getTableColumns(metadata, expectedTable.Name, "user") // todo: read role from config
		if err != nil {
			t.Fatalf("Erro with searching expectedTable in metadata: %e", err)
		}

		if !elementsMatch(expectedTable, metadataTable) {
			t.Errorf("Table columns do not match: %s", expectedTable.Name)
		}
	}
}

func elementsMatch(expectedTable hasura.ExpectedTable, metadataTable hasura.Columns) bool {
	if len(expectedTable.Columns) != len(metadataTable) {
		return false
	}

	hasuraColumns := make(map[string]int)

	for _, columnName := range metadataTable {
		hasuraColumns[columnName] = 0
	}

	for _, expectedColumn := range expectedTable.Columns {
		if _, ok := hasuraColumns[expectedColumn]; !ok {
			return false
		}
	}

	return true
}

func getTableColumns(metadata hasura.Metadata, tableName string, role string) (hasura.Columns, error) {
	for _, source := range metadata.Sources {
		for _, table := range source.Tables {
			if table.Schema.Name == tableName {
				for _, selectPermission := range table.SelectPermissions {
					if selectPermission.Role == role {
						return selectPermission.Permission.Columns, nil
					}
				}
			}
		}
	}

	return nil, errors.Errorf("Table %s for role %s was not found", tableName, role)
}
