package services_test

import (
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

// TestStartingPointService_SaveCapitalAssetStep_RejectsInvalidOnComplete is
// the representative case for the fix to "wizard hub services never call
// their own domain validators before persisting a complete item"
// (AGENTS.md's DDD violations audit): SaveCapitalAssetStep must reject an
// invalid CapitalAsset when the caller asks for StatusComplete, even though
// the equivalent web handler no longer runs domain.ValidateCapitalAsset
// itself - the service is now the authoritative gate.
func TestStartingPointService_SaveCapitalAssetStep_RejectsInvalidOnComplete(t *testing.T) {
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
	if err := svc.SaveCapitalAssetStep(itemID, invalidAsset, 1, repositories.StatusDraft); err != nil {
		t.Fatalf("expected draft save to succeed without validation, got error: %v", err)
	}

	// The finish step (StatusComplete) must reject it.
	if err := svc.SaveCapitalAssetStep(itemID, invalidAsset, 4, repositories.StatusComplete); err == nil {
		t.Error("expected SaveCapitalAssetStep to reject an invalid asset on StatusComplete, got nil")
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
