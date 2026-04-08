package migrations

import "testing"

func TestMigrationsAreValid(t *testing.T) {
	if err := validate(list()); err != nil {
		t.Fatalf("validate(list()) error: %v", err)
	}
}
