// Package handlers - Starting Point wizard: a summary page
// (/plan/{id}/starting-point) listing each section's completion status,
// plus one true multi-step wizard per section. Each step is its own saved
// HTTP round-trip so a user can leave mid-section and resume exactly where
// they left off.
package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// Ordered step names per section. Fixed Assets' "useful-life" step is
// skipped when the user picks DepreciationMethod=None - handled as a
// branch inside PostFixedAssetStep rather than a second step list, since
// it's the only section whose steps depend on data the user just entered.
var (
	fixedAssetSteps    = []string{"name", "cost", "depreciation-method", "useful-life"}
	startupCostSteps   = []string{"name", "amount"}
	fundingSourceSteps = []string{"name", "amount", "interest-rate", "term"}
	cashOnHandSteps    = []string{"cash", "accounts-receivable", "prepaid-expenses", "accounts-payable", "accrued-expenses"}
)

// startingPointComplete reports whether every Starting Point sub-section is
// complete for planID. Used to drive the sidebar's Starting Point nav icon,
// which is shown on every page, not just Starting Point's own.
func (app *App) startingPointComplete(planID uuid.UUID) bool {
	return app.HubSvc.Get(planID).StartingPoint
}

func stepIndex(steps []string, current string) int {
	for i, s := range steps {
		if s == current {
			return i
		}
	}
	return -1
}

func nextStepName(steps []string, idx int) string {
	if idx+1 >= len(steps) {
		return ""
	}
	return steps[idx+1]
}

func prevStepName(steps []string, idx int) string {
	if idx <= 0 {
		return ""
	}
	return steps[idx-1]
}

// cashOnHandButtonLabel is "Finish" on the last of Cash on Hand's 5 fixed
// steps (it isn't repeatable, so there's no "Add another?" detour after
// it) and "Next" everywhere else.
func cashOnHandButtonLabel(idx int) string {
	if idx == len(cashOnHandSteps)-1 {
		return "Finish"
	}
	return "Next"
}

// parseStepMoney parses a required, non-negative money field submitted by
// a single wizard step.
func parseStepMoney(raw string) (domain.Money, bool) {
	val, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || val < 0 {
		return 0, false
	}
	return domain.Money(val), true
}

// isStartingPointSection reports whether section is one of the 4 known
// Starting Point sections, guarding the {section} wildcard in
// GetStartingPointSectionIntro's route against arbitrary path input.
func isStartingPointSection(section string) bool {
	switch section {
	case domain.SectionFixedAssets, domain.SectionStartupCosts, domain.SectionFundingSources, domain.SectionCashOnHand:
		return true
	default:
		return false
	}
}

// GetStartingPointSectionIntro GET /plan/{id}/starting-point/{section}/intro
// Shows a brief description of what the section covers, with a single
// "Let's Get Started" action into the section's actual entry point. Shown
// the first time a user clicks "Get Started" from the summary page; the
// "Edit" link on an already-started section skips this and goes straight
// to the section itself.
func (app *App) GetStartingPointSectionIntro() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		section := r.PathValue("section")
		if !isStartingPointSection(section) {
			app.renderErrorPage(w, r, http.StatusNotFound, "That Starting Point section doesn't exist.")
			return
		}

		page := views.BuildStartingPointSectionIntroPage(r, user, plan, section, app.startingPointComplete(planID))
		views.RenderSectionIntroPage(w, app.TemplateCache, page)
	}
}

// GetStartingPointSummary GET /plan/{id}/starting-point
func (app *App) GetStartingPointSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "The Plan ID in the URL is malformed or invalid.")
			return
		}

		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that business plan. It may have been deleted, or you might be using an old link.")
			return
		}

		sectionStatus, err := app.StartingPointSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
			sectionStatus = map[string]bool{}
		}

		fixedAssets, err := app.StartingPointSvc.ListCompleteCapitalAssets(planID)
		if err != nil {
			log.Printf("Failed to load capital assets for plan %s: %v", planID, err)
		}
		startupCosts, err := app.StartingPointSvc.ListCompleteStartupCosts(planID)
		if err != nil {
			log.Printf("Failed to load startup costs for plan %s: %v", planID, err)
		}
		fundingSources, err := app.StartingPointSvc.ListCompleteFundingSources(planID)
		if err != nil {
			log.Printf("Failed to load funding sources for plan %s: %v", planID, err)
		}

		page := views.BuildStartingPointSummaryPage(r, user, plan, sectionStatus, len(fixedAssets), len(startupCosts), len(fundingSources))
		views.RenderHubSummaryPage(w, app.TemplateCache, page)
	}
}

// --- Fixed Assets ---

// GetFixedAssetList GET /plan/{id}/starting-point/fixed-assets
func (app *App) GetFixedAssetList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		completeItems, err := app.StartingPointSvc.ListCompleteCapitalAssets(planID)
		if err != nil {
			log.Printf("Failed to load capital assets for plan %s: %v", planID, err)
		}
		items := make([]views.StartingPointFixedAssetItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.StartingPointFixedAssetItem{ID: it.ID, Asset: it.Asset}
		}

		sectionStatus, err := app.StartingPointSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.StartingPointSvc.GetCapitalAssetDraft(planID); err != nil {
			log.Printf("Failed to load capital asset draft for plan %s: %v", planID, err)
		} else if draft != nil {
			id := draft.ID
			draftItemID = &id
			idx := draft.CurrentStep
			if idx < 0 || idx >= len(fixedAssetSteps) {
				idx = 0
			}
			draftStep = fixedAssetSteps[idx]
		}

		page := views.BuildFixedAssetsListPage(r, user, plan, items, sectionStatus[domain.SectionFixedAssets], draftItemID, draftStep, views.IsStartingPointComplete(sectionStatus))
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostFixedAssetNew POST /plan/{id}/starting-point/fixed-assets/new
// Starts a new Fixed Asset item, or resumes the plan's existing draft if
// one is already in progress.
func (app *App) PostFixedAssetNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.StartingPointSvc.CreateCapitalAssetDraft(planID)
		if err != nil {
			log.Printf("Failed to create capital asset draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new fixed asset. Please try again.")
			return
		}

		http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionFixedAssets, fixedAssetSteps[0]), http.StatusSeeOther)
	}
}

// GetFixedAssetStep GET /plan/{id}/starting-point/fixed-assets/{itemID}/{step}
func (app *App) GetFixedAssetStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(fixedAssetSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.StartingPointSvc.GetCapitalAsset(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(fixedAssetSteps, idx); prev != "" {
			backURL = views.SectionStepURL(planID, itemID, domain.SectionFixedAssets, prev)
		}

		page := views.BuildFixedAssetStepPage(r, user, plan, itemID, item.Asset, step, idx+1, len(fixedAssetSteps), backURL, "Next", "", app.startingPointComplete(planID))
		views.RenderFixedAssetStepPage(w, app.TemplateCache, page)
	}
}

// PostFixedAssetStep POST /plan/{id}/starting-point/fixed-assets/{itemID}/{step}
func (app *App) PostFixedAssetStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(fixedAssetSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.StartingPointSvc.GetCapitalAsset(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		asset := item.Asset
		wasComplete := item.Status == repositories.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(fixedAssetSteps, idx); prev != "" {
				backURL = views.SectionStepURL(planID, itemID, domain.SectionFixedAssets, prev)
			}
			page := views.BuildFixedAssetStepPage(r, user, plan, itemID, asset, step, idx+1, len(fixedAssetSteps), backURL, "Next", errMsg, app.startingPointComplete(planID))
			views.RenderFixedAssetStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
		}

		switch step {
		case "name":
			asset.Name = strings.TrimSpace(r.PostForm.Get("name"))
			if asset.Name == "" {
				renderStepError("Please enter a name for this asset.")
				return
			}
		case "cost":
			cost, ok := parseStepMoney(r.PostForm.Get("cost"))
			if !ok {
				renderStepError("Please enter a valid purchase cost.")
				return
			}
			asset.PurchaseCost = cost
		case "depreciation-method":
			method := domain.DepreciationMethod(r.PostForm.Get("depreciation_method"))
			if method != domain.StraightLine && method != domain.DoubleDeclining && method != domain.None {
				renderStepError("Please choose a depreciation method.")
				return
			}
			asset.DepreciationMethod = method
			if method == domain.None {
				asset.UsefulLifeMonths = 0
			}
		case "useful-life":
			years, err := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("useful_life_years")))
			if err != nil || years <= 0 {
				years = 5 // smart default, matches the previous single-page form's behavior
			}
			asset.UsefulLifeMonths = years * 12
		}

		// Land doesn't depreciate: a None method finishes the item right
		// after this step, skipping "useful-life" entirely.
		finishNow := step == "useful-life" || (step == "depreciation-method" && asset.DepreciationMethod == domain.None)

		// Full cross-field validation only makes sense once every field has
		// actually been answered - running it after every intermediate step
		// spuriously rejects a fresh draft (e.g. DepreciationMethod is still
		// "" on the "name"/"cost" steps, tripping the "must have a useful
		// life unless None" rule before the user has even reached that
		// question).
		if finishNow {
			if err := domain.ValidateCapitalAsset(asset); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := repositories.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = repositories.StatusComplete
			newCurrentStep = len(fixedAssetSteps)
		}

		if err := app.StartingPointSvc.SaveCapitalAssetStep(itemID, asset, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save fixed asset step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if !finishNow {
			http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionFixedAssets, nextStepName(fixedAssetSteps, idx)), http.StatusSeeOther)
			return
		}

		if wasComplete {
			// Editing an already-finished item: no "Add another?" detour.
			http.Redirect(w, r, views.SectionListURL(planID, domain.SectionFixedAssets), http.StatusSeeOther)
			return
		}

		page := views.BuildFixedAssetsAddAnotherPage(r, user, plan, asset, app.startingPointComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostFixedAssetDelete POST /plan/{id}/starting-point/fixed-assets/{itemID}
func (app *App) PostFixedAssetDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		if err := app.StartingPointSvc.DeleteCapitalAsset(itemID); err != nil {
			log.Printf("Failed to delete capital asset %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.SectionListURL(planID, domain.SectionFixedAssets), http.StatusSeeOther)
	}
}

// PostFixedAssetFinish POST /plan/{id}/starting-point/fixed-assets/finish
func (app *App) PostFixedAssetFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.StartingPointSvc.MarkWizardSectionComplete(planID, domain.SectionFixedAssets); err != nil {
			log.Printf("Failed to mark fixed assets complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
	}
}

// --- Startup Costs ---

// GetStartupCostList GET /plan/{id}/starting-point/startup-costs
func (app *App) GetStartupCostList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		completeItems, err := app.StartingPointSvc.ListCompleteStartupCosts(planID)
		if err != nil {
			log.Printf("Failed to load startup costs for plan %s: %v", planID, err)
		}
		items := make([]views.StartingPointStartupCostItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.StartingPointStartupCostItem{ID: it.ID, Cost: it.Cost}
		}

		sectionStatus, err := app.StartingPointSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.StartingPointSvc.GetStartupCostDraft(planID); err != nil {
			log.Printf("Failed to load startup cost draft for plan %s: %v", planID, err)
		} else if draft != nil {
			id := draft.ID
			draftItemID = &id
			idx := draft.CurrentStep
			if idx < 0 || idx >= len(startupCostSteps) {
				idx = 0
			}
			draftStep = startupCostSteps[idx]
		}

		page := views.BuildStartupCostsListPage(r, user, plan, items, sectionStatus[domain.SectionStartupCosts], draftItemID, draftStep, views.IsStartingPointComplete(sectionStatus))
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostStartupCostNew POST /plan/{id}/starting-point/startup-costs/new
func (app *App) PostStartupCostNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.StartingPointSvc.CreateStartupCostDraft(planID)
		if err != nil {
			log.Printf("Failed to create startup cost draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new startup cost. Please try again.")
			return
		}

		http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionStartupCosts, startupCostSteps[0]), http.StatusSeeOther)
	}
}

// GetStartupCostStep GET /plan/{id}/starting-point/startup-costs/{itemID}/{step}
func (app *App) GetStartupCostStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(startupCostSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.StartingPointSvc.GetStartupCost(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(startupCostSteps, idx); prev != "" {
			backURL = views.SectionStepURL(planID, itemID, domain.SectionStartupCosts, prev)
		}

		page := views.BuildStartupCostStepPage(r, user, plan, itemID, item.Cost, step, idx+1, len(startupCostSteps), backURL, "Next", "", app.startingPointComplete(planID))
		views.RenderStartupCostStepPage(w, app.TemplateCache, page)
	}
}

// PostStartupCostStep POST /plan/{id}/starting-point/startup-costs/{itemID}/{step}
func (app *App) PostStartupCostStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(startupCostSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.StartingPointSvc.GetStartupCost(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		cost := item.Cost
		wasComplete := item.Status == repositories.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(startupCostSteps, idx); prev != "" {
				backURL = views.SectionStepURL(planID, itemID, domain.SectionStartupCosts, prev)
			}
			page := views.BuildStartupCostStepPage(r, user, plan, itemID, cost, step, idx+1, len(startupCostSteps), backURL, "Next", errMsg, app.startingPointComplete(planID))
			views.RenderStartupCostStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
		}

		switch step {
		case "name":
			cost.Name = strings.TrimSpace(r.PostForm.Get("name"))
			if cost.Name == "" {
				renderStepError("Please enter a name for this startup cost.")
				return
			}
		case "amount":
			amount, ok := parseStepMoney(r.PostForm.Get("amount"))
			if !ok {
				renderStepError("Please enter a valid amount.")
				return
			}
			cost.Amount = amount
		}

		finishNow := idx == len(startupCostSteps)-1

		// Full validation only once every field has actually been
		// answered - see the matching comment in PostFixedAssetStep.
		if finishNow {
			if err := domain.ValidateStartupCost(cost); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := repositories.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = repositories.StatusComplete
		}

		if err := app.StartingPointSvc.SaveStartupCostStep(itemID, cost, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save startup cost step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if !finishNow {
			http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionStartupCosts, nextStepName(startupCostSteps, idx)), http.StatusSeeOther)
			return
		}

		if wasComplete {
			http.Redirect(w, r, views.SectionListURL(planID, domain.SectionStartupCosts), http.StatusSeeOther)
			return
		}

		page := views.BuildStartupCostsAddAnotherPage(r, user, plan, cost, app.startingPointComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostStartupCostDelete POST /plan/{id}/starting-point/startup-costs/{itemID}
func (app *App) PostStartupCostDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		if err := app.StartingPointSvc.DeleteStartupCost(itemID); err != nil {
			log.Printf("Failed to delete startup cost %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.SectionListURL(planID, domain.SectionStartupCosts), http.StatusSeeOther)
	}
}

// PostStartupCostFinish POST /plan/{id}/starting-point/startup-costs/finish
func (app *App) PostStartupCostFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.StartingPointSvc.MarkWizardSectionComplete(planID, domain.SectionStartupCosts); err != nil {
			log.Printf("Failed to mark startup costs complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
	}
}

// --- Funding Sources ---

// GetFundingSourceList GET /plan/{id}/starting-point/funding-sources
func (app *App) GetFundingSourceList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}

		completeItems, err := app.StartingPointSvc.ListCompleteFundingSources(planID)
		if err != nil {
			log.Printf("Failed to load funding sources for plan %s: %v", planID, err)
		}
		items := make([]views.StartingPointFundingSourceItem, len(completeItems))
		for i, it := range completeItems {
			items[i] = views.StartingPointFundingSourceItem{ID: it.ID, Funding: it.Funding}
		}

		sectionStatus, err := app.StartingPointSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
		}

		var draftItemID *uuid.UUID
		var draftStep string
		if draft, err := app.StartingPointSvc.GetFundingSourceDraft(planID); err != nil {
			log.Printf("Failed to load funding source draft for plan %s: %v", planID, err)
		} else if draft != nil {
			id := draft.ID
			draftItemID = &id
			idx := draft.CurrentStep
			if idx < 0 || idx >= len(fundingSourceSteps) {
				idx = 0
			}
			draftStep = fundingSourceSteps[idx]
		}

		page := views.BuildFundingSourcesListPage(r, user, plan, items, sectionStatus[domain.SectionFundingSources], draftItemID, draftStep, views.IsStartingPointComplete(sectionStatus))
		views.RenderSectionListPage(w, app.TemplateCache, page)
	}
}

// PostFundingSourceNew POST /plan/{id}/starting-point/funding-sources/new
func (app *App) PostFundingSourceNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		itemID, err := app.StartingPointSvc.CreateFundingSourceDraft(planID)
		if err != nil {
			log.Printf("Failed to create funding source draft for plan %s: %v", planID, err)
			app.renderErrorPage(w, r, http.StatusInternalServerError, "Failed to start a new funding source. Please try again.")
			return
		}

		http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionFundingSources, fundingSourceSteps[0]), http.StatusSeeOther)
	}
}

// GetFundingSourceStep GET /plan/{id}/starting-point/funding-sources/{itemID}/{step}
func (app *App) GetFundingSourceStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(fundingSourceSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.StartingPointSvc.GetFundingSource(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}

		backURL := ""
		if prev := prevStepName(fundingSourceSteps, idx); prev != "" {
			backURL = views.SectionStepURL(planID, itemID, domain.SectionFundingSources, prev)
		}

		page := views.BuildFundingSourceStepPage(r, user, plan, itemID, item.Funding, step, idx+1, len(fundingSourceSteps), backURL, "Next", "", app.startingPointComplete(planID))
		views.RenderFundingSourceStepPage(w, app.TemplateCache, page)
	}
}

// PostFundingSourceStep POST /plan/{id}/starting-point/funding-sources/{itemID}/{step}
func (app *App) PostFundingSourceStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(fundingSourceSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		item, err := app.StartingPointSvc.GetFundingSource(itemID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
			return
		}
		funding := item.Funding
		wasComplete := item.Status == repositories.StatusComplete

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(fundingSourceSteps, idx); prev != "" {
				backURL = views.SectionStepURL(planID, itemID, domain.SectionFundingSources, prev)
			}
			page := views.BuildFundingSourceStepPage(r, user, plan, itemID, funding, step, idx+1, len(fundingSourceSteps), backURL, "Next", errMsg, app.startingPointComplete(planID))
			views.RenderFundingSourceStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
		}

		switch step {
		case "name":
			funding.Name = strings.TrimSpace(r.PostForm.Get("name"))
			if funding.Name == "" {
				renderStepError("Please enter a name for this funding source.")
				return
			}
		case "amount":
			amount, ok := parseStepMoney(r.PostForm.Get("amount"))
			if !ok {
				renderStepError("Please enter a valid amount.")
				return
			}
			funding.Amount = amount
		case "interest-rate":
			rate, err := strconv.ParseFloat(strings.TrimSpace(r.PostForm.Get("interest_rate")), 64)
			if err != nil || rate < 0 {
				renderStepError("Please enter a valid interest rate (0 if this isn't a loan).")
				return
			}
			funding.InterestRate = rate / 100.0
		case "term":
			term, err := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("term_months")))
			if err != nil || term < 0 {
				renderStepError("Please enter a valid term in months (0 if this isn't a loan).")
				return
			}
			funding.TermMonths = term
		}

		finishNow := idx == len(fundingSourceSteps)-1

		// Full validation only once every field has actually been
		// answered - see the matching comment in PostFixedAssetStep.
		if finishNow {
			if err := domain.ValidateFundingSource(funding); err != nil {
				renderStepError(err.Error())
				return
			}
		}

		newStatus := repositories.StatusDraft
		newCurrentStep := idx + 1
		if finishNow {
			newStatus = repositories.StatusComplete
		}

		if err := app.StartingPointSvc.SaveFundingSourceStep(itemID, funding, newCurrentStep, newStatus); err != nil {
			log.Printf("Failed to save funding source step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if !finishNow {
			http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionFundingSources, nextStepName(fundingSourceSteps, idx)), http.StatusSeeOther)
			return
		}

		if wasComplete {
			http.Redirect(w, r, views.SectionListURL(planID, domain.SectionFundingSources), http.StatusSeeOther)
			return
		}

		page := views.BuildFundingSourcesAddAnotherPage(r, user, plan, funding, app.startingPointComplete(planID))
		views.RenderAddAnotherPage(w, app.TemplateCache, page)
	}
}

// PostFundingSourceDelete POST /plan/{id}/starting-point/funding-sources/{itemID}
func (app *App) PostFundingSourceDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid item ID")
			return
		}
		if err := app.StartingPointSvc.DeleteFundingSource(itemID); err != nil {
			log.Printf("Failed to delete funding source %s: %v", itemID, err)
		}
		http.Redirect(w, r, views.SectionListURL(planID, domain.SectionFundingSources), http.StatusSeeOther)
	}
}

// PostFundingSourceFinish POST /plan/{id}/starting-point/funding-sources/finish
func (app *App) PostFundingSourceFinish() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		if err := app.StartingPointSvc.MarkWizardSectionComplete(planID, domain.SectionFundingSources); err != nil {
			log.Printf("Failed to mark funding sources complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
	}
}

// --- Cash on Hand (singleton per plan, always exactly 5 steps) ---

// cashOnHandFieldSet writes a validated new value into the StartingBalances
// field a given step answers.
func cashOnHandFieldSet(balances domain.StartingBalances, step string, amount domain.Money) domain.StartingBalances {
	switch step {
	case "cash":
		balances.Cash = amount
	case "accounts-receivable":
		balances.AccountsReceivable = amount
	case "prepaid-expenses":
		balances.PrepaidExpenses = amount
	case "accounts-payable":
		balances.AccountsPayable = amount
	case "accrued-expenses":
		balances.AccruedExpenses = amount
	}
	return balances
}

// GetCashOnHandEntry GET /plan/{id}/starting-point/cash-on-hand
// Redirects into the wizard at the right step: the top, for a plan that
// has never completed Cash on Hand (or wants to re-answer it), or wherever
// the user left off if they abandoned it mid-flow.
func (app *App) GetCashOnHandEntry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}

		sectionStatus, err := app.StartingPointSvc.GetHubStatus(planID)
		if err != nil {
			log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
		}

		step := cashOnHandSteps[0]
		if !sectionStatus[domain.SectionCashOnHand] {
			if row, err := app.StartingPointSvc.GetStartingBalancesRow(planID); err != nil {
				log.Printf("Failed to load starting balances for plan %s: %v", planID, err)
			} else if row.CurrentStep > 0 && row.CurrentStep < len(cashOnHandSteps) {
				step = cashOnHandSteps[row.CurrentStep]
			}
		}

		http.Redirect(w, r, views.SectionSingletonStepURL(planID, domain.SectionCashOnHand, step), http.StatusSeeOther)
	}
}

// GetCashOnHandStep GET /plan/{id}/starting-point/cash-on-hand/{step}
func (app *App) GetCashOnHandStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(cashOnHandSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		row, err := app.StartingPointSvc.GetStartingBalancesRow(planID)
		if err != nil {
			log.Printf("Failed to load starting balances for plan %s: %v", planID, err)
			row = &repositories.StartingBalancesRow{}
		}

		backURL := ""
		if prev := prevStepName(cashOnHandSteps, idx); prev != "" {
			backURL = views.SectionSingletonStepURL(planID, domain.SectionCashOnHand, prev)
		}

		page := views.BuildCashOnHandStepPage(r, user, plan, row.Balances, step, idx+1, len(cashOnHandSteps), backURL, cashOnHandButtonLabel(idx), "", app.startingPointComplete(planID))
		views.RenderCashOnHandStepPage(w, app.TemplateCache, page)
	}
}

// PostCashOnHandStep POST /plan/{id}/starting-point/cash-on-hand/{step}
func (app *App) PostCashOnHandStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		planID, err := parsePlanID(r)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID")
			return
		}
		plan, err := app.PlanSvc.Get(planID)
		if err != nil {
			app.renderErrorPage(w, r, http.StatusNotFound, "Plan not found")
			return
		}
		step := r.PathValue("step")
		idx := stepIndex(cashOnHandSteps, step)
		if idx == -1 {
			app.renderErrorPage(w, r, http.StatusNotFound, "That step doesn't exist.")
			return
		}

		row, err := app.StartingPointSvc.GetStartingBalancesRow(planID)
		if err != nil {
			log.Printf("Failed to load starting balances for plan %s: %v", planID, err)
			row = &repositories.StartingBalancesRow{}
		}
		balances := row.Balances

		renderStepError := func(errMsg string) {
			backURL := ""
			if prev := prevStepName(cashOnHandSteps, idx); prev != "" {
				backURL = views.SectionSingletonStepURL(planID, domain.SectionCashOnHand, prev)
			}
			page := views.BuildCashOnHandStepPage(r, user, plan, balances, step, idx+1, len(cashOnHandSteps), backURL, cashOnHandButtonLabel(idx), errMsg, app.startingPointComplete(planID))
			views.RenderCashOnHandStepPageWithStatus(w, app.TemplateCache, page, http.StatusBadRequest)
		}

		amount, ok := parseStepMoney(r.PostForm.Get("amount"))
		if !ok {
			renderStepError("Please enter a valid amount.")
			return
		}
		balances = cashOnHandFieldSet(balances, step, amount)

		newCurrentStep := idx + 1
		if err := app.StartingPointSvc.SaveStartingBalancesStep(planID, balances, newCurrentStep); err != nil {
			log.Printf("Failed to save starting balances step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
			return
		}

		if idx < len(cashOnHandSteps)-1 {
			http.Redirect(w, r, views.SectionSingletonStepURL(planID, domain.SectionCashOnHand, nextStepName(cashOnHandSteps, idx)), http.StatusSeeOther)
			return
		}

		// Last step: Cash on Hand isn't repeatable, so finishing it marks
		// the section complete directly and returns to the summary page.
		if err := app.StartingPointSvc.MarkWizardSectionComplete(planID, domain.SectionCashOnHand); err != nil {
			log.Printf("Failed to mark cash on hand complete for plan %s: %v", planID, err)
		}
		http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
	}
}
