package main

import (
	"errors"
	"testing"
)

func TestMigrationHintPermissionDenied(t *testing.T) {
	t.Parallel()
	err := errors.New(`the following errors occurred:
 -  ERROR: relation "goose_db_version" does not exist (SQLSTATE 42P01)
 -  ERROR: permission denied for schema public (SQLSTATE 42501)`)
	got := migrationHint(err)
	if got == "" {
		t.Fatal("expected a DigitalOcean grant hint")
	}
	if migrationHint(errors.New("connection refused")) != "" {
		t.Fatal("unrelated errors must not hint grants")
	}
}
