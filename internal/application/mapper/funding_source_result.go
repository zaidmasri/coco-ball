package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewFundingSourceResultFromEntity builds a FundingSourceResult from a
// FundingSource value object.
func NewFundingSourceResultFromEntity(f domain.FundingSource) *common.FundingSourceResult {
	return &common.FundingSourceResult{
		ID:           f.ID(),
		Name:         f.Name,
		Amount:       f.Amount,
		InterestRate: f.InterestRate,
		TermMonths:   f.TermMonths,
	}
}
