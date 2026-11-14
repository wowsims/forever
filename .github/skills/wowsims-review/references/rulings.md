# Rulings

Each bullet is one rule the code follows: its id, the rule, where it lives, the pull request that settled it, and the
evidence that measured it (`docs/evidence/`). An id is never reused; a replaced rule keeps its bullet with a
`Replaced by` note.

- **R-0049** **A helper, mask or empty body the change itself orphaned goes in the same change; a generic helper or empty
  extension table kept for a named upcoming use (core helpers, override maps, icon override tables, phase ladders such
  as `InferPhase`) stays, with a comment naming that use, and anything generic is asked about before removal.**
  `tools/database/item_source_utils.go`, PR #215.
