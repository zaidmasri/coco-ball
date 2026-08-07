package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type ProductRepository interface {
	CreateProductDraft(planID uuid.UUID) (uuid.UUID, error)
	GetProductDraft(planID uuid.UUID) (*ProductItem, error)
	GetProduct(itemID uuid.UUID) (*ProductItem, error)
	SaveProductStep(itemID uuid.UUID, product domain.Product, currentStep int, status ItemStatus) error
	ListCompleteProducts(planID uuid.UUID) ([]ProductItem, error)
	DeleteProduct(itemID uuid.UUID) error
}
