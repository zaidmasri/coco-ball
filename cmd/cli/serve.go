package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"time"

	appservices "github.com/zaidmasri/business-planning-tool/internal/application/services"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
	"github.com/zaidmasri/business-planning-tool/internal/interface/web"
	"github.com/zaidmasri/business-planning-tool/internal/middleware"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

func serve(dbPath, port string) {
	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer conn.Close()

	if err := sqlite.RunMigrations(conn); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Drains the outbox table written by aggregate repositories on every save.
	relay := sqlite.NewOutboxRelay(conn, 2*time.Second)
	relay.Start(context.Background())

	// Infrastructure: one sqlc-backed repository per domain interface.
	planRepo := sqlite.NewPlanRepository(conn)
	userRepo := sqlite.NewUserRepository(conn)
	sessionRepo := sqlite.NewSessionRepository(conn)
	accessRepo := sqlite.NewAccessRepository(conn)
	inviteRepo := sqlite.NewInviteRepository(conn)
	wizardProgressRepo := sqlite.NewWizardProgressRepository(conn)
	capitalAssetRepo := sqlite.NewCapitalAssetRepository(conn)
	startupCostRepo := sqlite.NewStartupCostRepository(conn)
	fundingSourceRepo := sqlite.NewFundingSourceRepository(conn)
	startingBalancesRepo := sqlite.NewStartingBalancesRepository(conn)
	salaryRoleRepo := sqlite.NewSalaryRoleRepository(conn)
	benefitRepo := sqlite.NewBenefitRepository(conn)
	payrollTaxRatesRepo := sqlite.NewPayrollTaxRatesRepository(conn)
	productRepo := sqlite.NewProductRepository(conn)
	salesGrowthCurveRepo := sqlite.NewSalesGrowthCurveRepository(conn)
	inventoryPurchaseRepo := sqlite.NewInventoryPurchaseRepository(conn)
	distributionRepo := sqlite.NewDistributionRepository(conn)
	operatingExpenseRepo := sqlite.NewOperatingExpenseRepository(conn)

	// Application services
	planSvc := appservices.NewPlanService(planRepo, userRepo)
	authSvc := appservices.NewAuthService(userRepo, sessionRepo)
	accessSvc := appservices.NewAccessService(accessRepo, planRepo, userRepo)
	inviteSvc := appservices.NewInviteService(inviteRepo, planRepo)
	startingPointSvc := appservices.NewStartingPointService(capitalAssetRepo, startupCostRepo, fundingSourceRepo, startingBalancesRepo, wizardProgressRepo)
	payrollSvc := appservices.NewPayrollService(salaryRoleRepo, benefitRepo, payrollTaxRatesRepo, wizardProgressRepo)
	salesForecastSvc := appservices.NewSalesForecastService(productRepo, salesGrowthCurveRepo, wizardProgressRepo)
	cashFlowSvc := appservices.NewCashFlowService(inventoryPurchaseRepo, distributionRepo, wizardProgressRepo)
	opExSvc := appservices.NewOperatingExpensesService(operatingExpenseRepo, wizardProgressRepo)
	hubSvc := appservices.NewHubCompletionService(startingPointSvc, payrollSvc, salesForecastSvc, cashFlowSvc, opExSvc)

	// Load templates
	templateCache := views.LoadTemplates()

	mux := http.NewServeMux()

	// Static files
	staticFS, err := fs.Sub(views.StaticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static fs: %v", err)
	}
	staticFileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFileServer))

	// Cross-cutting middleware shared by the controllers below.
	accessMW := web.NewPlanAccessMiddleware(accessSvc, templateCache)
	sessionMW := web.NewSessionMiddleware(authSvc)

	// HTTP layer: one self-registering controller per domain.
	web.NewAuthController(mux, authSvc, templateCache)
	web.NewPlanController(mux, planSvc, accessSvc, inviteSvc, authSvc, hubSvc, templateCache, accessMW)
	web.NewInviteController(mux, inviteSvc, planSvc, accessSvc, hubSvc, templateCache, accessMW)
	web.NewStartingPointController(mux, planSvc, startingPointSvc, templateCache, accessMW)
	web.NewPayrollController(mux, planSvc, payrollSvc, templateCache, accessMW)
	web.NewSalesForecastController(mux, planSvc, salesForecastSvc, templateCache, accessMW)
	web.NewCashFlowController(mux, planSvc, cashFlowSvc, templateCache, accessMW)
	web.NewOperatingExpensesController(mux, planSvc, opExSvc, templateCache, accessMW)

	mux.HandleFunc("/", web.NotFound(templateCache))

	// Wrap mux with middlewares
	var httpHandler http.Handler = mux
	httpHandler = sessionMW.Authenticate(httpHandler)
	httpHandler = middleware.Logger(httpHandler)
	httpHandler = web.Recover(templateCache)(httpHandler)

	// Configure a robust http server
	srv := &http.Server{
		Addr:         port,
		Handler:      httpHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on %s", port)
	log.Fatal(srv.ListenAndServe())
}
