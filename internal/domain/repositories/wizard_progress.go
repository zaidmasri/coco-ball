package repositories

import "github.com/google/uuid"

type WizardProgressRepository interface {
	MarkWizardSectionComplete(planID uuid.UUID, hub, section string) error
	GetWizardSectionStatus(planID uuid.UUID, hub string) (map[string]bool, error)
}
