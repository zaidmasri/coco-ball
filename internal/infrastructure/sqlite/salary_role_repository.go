package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// SalaryRoleRepository implements repositories.SalaryRoleRepository.
type SalaryRoleRepository struct {
	queries *db.Queries
}

func NewSalaryRoleRepository(conn *sql.DB) repositories.SalaryRoleRepository {
	return &SalaryRoleRepository{queries: db.New(conn)}
}

var _ repositories.SalaryRoleRepository = (*SalaryRoleRepository)(nil)

func (r *SalaryRoleRepository) CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindSalaryRoleDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxSalaryRoleSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.NewV7()

	now := time.Now().Unix()
	if err := r.queries.CreateSalaryRoleDraft(ctx, db.CreateSalaryRoleDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create salary role draft: %w", err)
	}

	return id, nil
}

func (r *SalaryRoleRepository) FindSalaryRoleDraft(planID uuid.UUID) (*repositories.SalaryRoleItem, error) {
	row, err := r.queries.FindSalaryRoleDraft(context.Background(), db.FindSalaryRoleDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find salary role draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse salary role id: %w", err)
	}

	return &repositories.SalaryRoleItem{
		ID:          id,
		Role:        salaryRoleFromRow(row.Role, row.IsContractor, row.Headcount, row.MonthlyPay, row.GrowthYr2, row.GrowthYr3),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *SalaryRoleRepository) GetSalaryRole(itemID uuid.UUID) (*repositories.SalaryRoleItem, error) {
	row, err := r.queries.GetSalaryRole(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("salary role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get salary role: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse salary role id: %w", err)
	}

	return &repositories.SalaryRoleItem{
		ID:          id,
		Role:        salaryRoleFromRow(row.Role, row.IsContractor, row.Headcount, row.MonthlyPay, row.GrowthYr2, row.GrowthYr3),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *SalaryRoleRepository) SaveSalaryRoleStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int, status repositories.ItemStatus) error {
	var yr2, yr3 float64
	if len(role.GrowthAfterYr1.RatesAfterYear1) > 0 {
		yr2 = role.GrowthAfterYr1.RatesAfterYear1[0]
	}
	if len(role.GrowthAfterYr1.RatesAfterYear1) > 1 {
		yr3 = role.GrowthAfterYr1.RatesAfterYear1[1]
	}

	affected, err := r.queries.SaveSalaryRoleStep(context.Background(), db.SaveSalaryRoleStepParams{
		Role:         role.Role,
		IsContractor: boolToInt64(role.IsContractor),
		Headcount:    int64(role.Headcount),
		MonthlyPay:   role.MonthlyPay.MinorUnits(),
		GrowthYr2:    yr2,
		GrowthYr3:    yr3,
		Status:       string(status),
		CurrentStep:  int64(currentStep),
		UpdatedAt:    time.Now().Unix(),
		ID:           itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save salary role step: %w", err)
	}
	if affected == 0 {
		return errors.New("salary role not found")
	}
	return nil
}

func (r *SalaryRoleRepository) ListCompleteSalaryRoles(planID uuid.UUID) ([]repositories.SalaryRoleItem, error) {
	rows, err := r.queries.ListCompleteSalaryRoles(context.Background(), db.ListCompleteSalaryRolesParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list salary roles: %w", err)
	}

	items := make([]repositories.SalaryRoleItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse salary role id: %w", err)
		}
		items[i] = repositories.SalaryRoleItem{
			ID:          id,
			Role:        salaryRoleFromRow(row.Role, row.IsContractor, row.Headcount, row.MonthlyPay, row.GrowthYr2, row.GrowthYr3),
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *SalaryRoleRepository) DeleteSalaryRole(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteSalaryRole(context.Background(), db.DeleteSalaryRoleParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete salary role: %w", err)
	}
	if affected == 0 {
		return errors.New("salary role not found")
	}
	return nil
}

func salaryRoleFromRow(role string, isContractor, headcount, monthlyPay int64, growthYr2, growthYr3 float64) domain.SalaryRole {
	return domain.SalaryRole{
		Role:           role,
		IsContractor:   int64ToBool(isContractor),
		Headcount:      int(headcount),
		MonthlyPay:     mustUSD(monthlyPay),
		GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
	}
}
