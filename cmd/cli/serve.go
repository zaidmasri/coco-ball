package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"time"

	appservices "github.com/zaidmasri/business-planning-tool/internal/application/services"
	"github.com/zaidmasri/business-planning-tool/internal/handlers"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
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
	planSvc := appservices.NewPlanService(planRepo)
	authSvc := appservices.NewAuthService(userRepo, sessionRepo)
	accessSvc := appservices.NewAccessService(accessRepo)
	inviteSvc := appservices.NewInviteService(inviteRepo)
	startingPointSvc := appservices.NewStartingPointService(capitalAssetRepo, startupCostRepo, fundingSourceRepo, startingBalancesRepo, wizardProgressRepo)
	payrollSvc := appservices.NewPayrollService(salaryRoleRepo, benefitRepo, payrollTaxRatesRepo, wizardProgressRepo)
	salesForecastSvc := appservices.NewSalesForecastService(productRepo, salesGrowthCurveRepo, wizardProgressRepo)
	cashFlowSvc := appservices.NewCashFlowService(inventoryPurchaseRepo, distributionRepo, wizardProgressRepo)
	opExSvc := appservices.NewOperatingExpensesService(operatingExpenseRepo, wizardProgressRepo)
	hubSvc := appservices.NewHubCompletionService(startingPointSvc, payrollSvc, salesForecastSvc, cashFlowSvc, opExSvc)

	// Load templates
	templateCache := loadTemplates()

	// HTTP layer
	app := handlers.NewApp(planSvc, authSvc, accessSvc, inviteSvc,
		startingPointSvc, payrollSvc, salesForecastSvc, cashFlowSvc, opExSvc, hubSvc,
		templateCache)

	mux := http.NewServeMux()

	// Static files
	staticFS, err := fs.Sub(views.StaticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static fs: %v", err)
	}
	staticFileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFileServer))

	// Register all application routes
	app.RegisterRoutes(mux)

	// Wrap mux with middlewares
	var httpHandler http.Handler = mux
	httpHandler = app.AuthMiddleware(httpHandler)
	httpHandler = middleware.Logger(httpHandler)

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

func loadTemplates() map[string]*template.Template {
	templateCache := make(map[string]*template.Template)

	pages, err := fs.Glob(views.TemplatesFS, "templates/pages/*.html")
	if err != nil {
		log.Fatal("error globbing templates:", err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		var ts *template.Template
		var parseErr error

		// Standalone pages that don't use base layout
		if name == "index.html" || name == "login.html" || name == "signup.html" || name == "profile.html" {
			ts, parseErr = template.New(name).Funcs(views.TemplateFuncs).ParseFS(views.TemplatesFS, page)
		} else {
			ts, parseErr = template.New("base.html").Funcs(views.TemplateFuncs).ParseFS(views.TemplatesFS, "templates/base.html", "templates/components/*.html", page)
		}

		if parseErr != nil {
			log.Fatalf("error parsing template %s: %v", name, parseErr)
		}

		templateCache[name] = ts
	}

	return templateCache
}
