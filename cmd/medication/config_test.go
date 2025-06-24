package main

import (
	"testing"
)

func TestConfig(t *testing.T) {
	t.Setenv("MED_LISTEN", ":42")
	t.Setenv("MED_DYNAMO_ENDPOINT", "http://localhost:8000")
	t.Setenv("MED_MEDICATION_TABLE", "my_table")

	c, err := NewConfig()
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if c.Listen != ":42" {
		t.Fatalf("invalid listen: %s", c.Listen)
	}
	if c.DynamoEndpoint != "http://localhost:8000" {
		t.Fatalf("invalid endpoint: %s", c.DynamoEndpoint)
	}
	if c.MedicationTable != "my_table" {
		t.Fatalf("invalid medication_table: %s", c.MedicationTable)
	}
}
