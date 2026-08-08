package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// ProductRepository implements repositories.ProductRepository.
type ProductRepository struct {
	queries *db.Queries
}

func NewProductRepository(conn *sql.DB) repositories.ProductRepository {
	return &ProductRepository{queries: db.New(conn)}
}

var _ repositories.ProductRepository = (*ProductRepository)(nil)

func (r *ProductRepository) CreateProductDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindProductDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxProductSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to generate id: %w", err)
	}

	now := time.Now().Unix()
	if err := r.queries.CreateProductDraft(ctx, db.CreateProductDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create product draft: %w", err)
	}

	return id, nil
}

func (r *ProductRepository) FindProductDraft(planID uuid.UUID) (*repositories.ProductItem, error) {
	row, err := r.queries.FindProductDraft(context.Background(), db.FindProductDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find product draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse product id: %w", err)
	}

	return &repositories.ProductItem{
		ID:          id,
		Product:     productFromRow(row.Name, row.Month1Units, row.PricePerUnit, row.CostPerUnit),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *ProductRepository) GetProduct(itemID uuid.UUID) (*repositories.ProductItem, error) {
	row, err := r.queries.GetProduct(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("product not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse product id: %w", err)
	}

	return &repositories.ProductItem{
		ID:          id,
		Product:     productFromRow(row.Name, row.Month1Units, row.PricePerUnit, row.CostPerUnit),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *ProductRepository) SaveProductStep(itemID uuid.UUID, product domain.Product, currentStep int, status repositories.ItemStatus) error {
	affected, err := r.queries.SaveProductStep(context.Background(), db.SaveProductStepParams{
		Name:         product.Name,
		Month1Units:  int64(product.Month1Units),
		PricePerUnit: product.PricePerUnit.MinorUnits(),
		CostPerUnit:  product.CostPerUnit.MinorUnits(),
		Status:       string(status),
		CurrentStep:  int64(currentStep),
		UpdatedAt:    time.Now().Unix(),
		ID:           itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save product step: %w", err)
	}
	if affected == 0 {
		return errors.New("product not found")
	}
	return nil
}

func (r *ProductRepository) ListCompleteProducts(planID uuid.UUID) ([]repositories.ProductItem, error) {
	rows, err := r.queries.ListCompleteProducts(context.Background(), db.ListCompleteProductsParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	items := make([]repositories.ProductItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse product id: %w", err)
		}
		items[i] = repositories.ProductItem{
			ID:          id,
			Product:     productFromRow(row.Name, row.Month1Units, row.PricePerUnit, row.CostPerUnit),
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *ProductRepository) DeleteProduct(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteProduct(context.Background(), db.DeleteProductParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if affected == 0 {
		return errors.New("product not found")
	}
	return nil
}

func productFromRow(name string, month1Units, pricePerUnit, costPerUnit int64) domain.Product {
	return domain.Product{
		Name:         name,
		Month1Units:  int(month1Units),
		PricePerUnit: mustUSD(pricePerUnit),
		CostPerUnit:  mustUSD(costPerUnit),
	}
}
