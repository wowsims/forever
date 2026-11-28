package database

// Rebuilds sim/core/dbcenums/forms_auto_gen.go from the SpellShapeshiftForm rows committed in
// assets/db_inputs/spell_store_inputs.json and asserts the committed file is what comes out. No
// game files database and no build tag.

import (
	"testing"
)

func TestFormsRegenerateFromTheCommittedInputs(t *testing.T) {
	inRepositoryRoot(t)

	rendered, err := renderFormsFile(committedInputs(t).Forms)
	if err != nil {
		t.Fatalf("rendering the forms file: %v", err)
	}
	assertRendersCommitted(t, "sim/core/dbcenums/forms_auto_gen.go", rendered)
}
