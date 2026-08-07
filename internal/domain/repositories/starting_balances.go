package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type StartingBalancesRepository interface {
	GetStartingBalancesRow(planID uuid.UUID) (*StartingBalancesRow, error)
	SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int) error
}
