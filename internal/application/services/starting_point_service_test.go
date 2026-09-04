package services_test

import (
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

// TestStartingPointService_CreateCapitalAsset_RejectsInvalid is the
// representative case for the fix to "wizard hub services never call their
// own domain validators before persisting a complete item" (AGENTS.md's DDD
// violations audit): CreateCapitalAsset must reject an invalid CapitalAsset,
// even though the equivalent web handler no longer runs
// domain.ValidateCapitalAsset itself - domain.NewCapitalAsset (called
// inside CreateCapitalAsset) is now the authoritative gate. SaveCapitalAssetDraftStep,
// used for the wizard's intermediate steps, must not validate - a fresh,
// incomplete draft is expected to be "invalid" by the full-entity rule.
func TestStartingPointService_CreateCapitalAsset_RejectsInvalid(t *testing.T) {
	plans, users, sessions, dbPath, cleanup := newTestStore(t)
	defer cleanup()

	ownerID := newPersistedOwner(t, users, sessions)
	planSvc := services.NewPlanService(plans, users)
	planResult, err := planSvc.CreatePlan(&commands.CreatePlan{
		Name:          "Starting Point Co",
		StartingMonth: 1,
		StartingYear:  2026,
		OwnerID:       ownerID,
	})
	if err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}

	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		t.Fatalf("NewConnection: %v", err)
	}
	defer conn.Close()
	capitalAssetRepo := sqlite.NewCapitalAssetRepository(conn)
	startupCostRepo := sqlite.NewStartupCostRepository(conn)
	fundingSourceRepo := sqlite.NewFundingSourceRepository(conn)
	startingBalancesRepo := sqlite.NewStartingBalancesRepository(conn)
	wizardProgressRepo := sqlite.NewWizardProgressRepository(conn)

	svc := services.NewStartingPointService(capitalAssetRepo, startupCostRepo, fundingSourceRepo, startingBalancesRepo, wizardProgressRepo)

	itemID, err := capitalAssetRepo.CreateCapitalAssetDraft(planResult.Result.ID)
	if err != nil {
		t.Fatalf("CreateCapitalAssetDraft: %v", err)
	}

	invalidAsset := domain.CapitalAsset{
		Name:               "Broken Asset",
		PurchaseCost:       mustUSDForTest(t, 100),
		SalvageValue:       mustUSDForTest(t, 500), // salvage > cost - invalid
		UsefulLifeMonths:   12,
		DepreciationMethod: domain.StraightLine,
	}

	// Draft saves (intermediate steps) must not validate - a fresh,
	// incomplete draft is expected to be "invalid" by the full-entity rule.
	if err := svc.SaveCapitalAssetDraftStep(itemID, invalidAsset, 1); err != nil {
		t.Fatalf("expected draft save to succeed without validation, got error: %v", err)
	}

	// CreateCapitalAsset (the finish step) must reject it.
	if _, err := svc.CreateCapitalAsset(&commands.CreateCapitalAsset{
		ItemID:             itemID,
		Name:               invalidAsset.Name,
		PurchaseCost:       invalidAsset.PurchaseCost,
		UsefulLifeMonths:   invalidAsset.UsefulLifeMonths,
		SalvageValue:       invalidAsset.SalvageValue,
		PurchaseMonthIndex: invalidAsset.PurchaseMonthIndex,
		DepreciationMethod: invalidAsset.DepreciationMethod,
		AssociatedLoan:     invalidAsset.AssociatedLoan,
		CurrentStep:        4,
	}); err == nil {
		t.Error("expected CreateCapitalAsset to reject an invalid asset, got nil")
	}
}

func mustUSDForTest(t *testing.T, dollars int64) domain.Money {
	t.Helper()
	m, err := domain.NewMoney(dollars*100, domain.USD)
	if err != nil {
		t.Fatalf("NewMoney: %v", err)
	}
	return m
}
