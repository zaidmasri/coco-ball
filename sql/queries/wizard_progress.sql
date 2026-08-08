-- name: UpsertWizardSection :exec
INSERT INTO wizard_sections (plan_id, hub, section, completed_at) VALUES (?, ?, ?, ?)
ON CONFLICT(plan_id, hub, section) DO UPDATE SET completed_at = excluded.completed_at;

-- name: ListWizardSections :many
SELECT section, completed_at FROM wizard_sections WHERE plan_id = ? AND hub = ?;
