package entities

import (
	"encoding/json"

	"github.com/google/uuid"
)

// planJSON is an intermediate struct for JSON marshalling.
//
// Starting Point, Payroll, Sales Forecast, Operating Expenses, and Cash
// Flow are deliberately NOT included here - they moved to normalized SQL
// tables (see internal/infrastructure/sqlite's plan_repository.go) so each
// wizard's sub-pages can save progressively without racing other sections.
// The repository layer populates those fields on Plan via each section's
// LoadXData method after unmarshalling everything else from this blob.
type planJSON struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	StartingMonth int             `json:"startingMonth"`
	StartingYear  int             `json:"startingYear"`
	OwnerID       string          `json:"ownerID"`
	Revenues      []RevenueStream `json:"revenues"`
	COGS          []Cost          `json:"cogs"`
}

// MarshalJSON implements json.Marshaler for Plan
func (p *Plan) MarshalJSON() ([]byte, error) {
	pj := planJSON{
		ID:            p.id.String(),
		Name:          p.name,
		StartingMonth: p.startingMonth,
		StartingYear:  p.startingYear,
		OwnerID:       p.ownerID.String(),
		Revenues:      p.revenues,
		COGS:          p.cogs,
	}

	return json.Marshal(pj)
}

// UnmarshalJSON implements json.Unmarshaler for Plan
func (p *Plan) UnmarshalJSON(data []byte) error {
	var pj planJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return err
	}

	id, err := uuid.Parse(pj.ID)
	if err != nil {
		return err
	}

	ownerID, err := uuid.Parse(pj.OwnerID)
	if err != nil {
		return err
	}

	p.id = id
	p.name = pj.Name
	p.startingMonth = pj.StartingMonth
	p.startingYear = pj.StartingYear
	p.ownerID = ownerID
	p.revenues = pj.Revenues
	p.cogs = pj.COGS

	return nil
}
