package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type ProductRepository interface {
	CreateProductDraft(planID uuid.UUID) (uuid.UUID, error)
	FindProductDraft(planID uuid.UUID) (*ProductItem, error)
	GetProduct(itemID uuid.UUID) (*ProductItem, error)
	SaveProductStep(itemID uuid.UUID, product domain.Product, currentStep int, status ItemStatus) error
	ListCompleteProducts(planID uuid.UUID) ([]ProductItem, error)
	DeleteProduct(itemID uuid.UUID) error
}
