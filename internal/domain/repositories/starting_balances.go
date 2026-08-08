package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type StartingBalancesRepository interface {
	GetStartingBalancesRow(planID uuid.UUID) (*StartingBalancesRow, error)
	SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int) error
}
