package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type ProductRepository interface {
	CreateProductDraft(planID uuid.UUID) (uuid.UUID, error)
	FindProductDraft(planID uuid.UUID) (*ProductItem, error)
	GetProduct(itemID uuid.UUID) (*ProductItem, error)
	// SaveProductDraftStep persists an in-progress, possibly incomplete
	// wizard step. Drafts are allowed to be invalid by design, so this
	// accepts a bare domain.Product.
	SaveProductDraftStep(itemID uuid.UUID, product domain.Product, currentStep int) error
	// CompleteProduct marks a product item complete. It requires a
	// domain.ValidatedProduct, so it is a compile error to complete an item
	// without validating it first - see domain.NewValidatedProduct.
	CompleteProduct(itemID uuid.UUID, product domain.ValidatedProduct, currentStep int) error
	ListCompleteProducts(planID uuid.UUID) ([]ProductItem, error)
	DeleteProduct(itemID uuid.UUID) error
}
