// Package web - Starting Point wizard: a summary page
// (/plan/{id}/starting-point) listing each section's completion status,
// plus one true multi-step wizard per section. Each step is its own saved
// HTTP round-trip so a user can leave mid-section and resume exactly where
// they left off.
package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"uuid"

	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
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

// StartingPointController handles the Starting Point hub: a summary page
// plus the Fixed Assets, Startup Costs, Funding Sources, and Cash on Hand
// sub-wizards.
type StartingPointController struct {
	planSvc          ifaces.PlanService
	startingPointSvc ifaces.StartingPointService
	templateCache    map[string]*template.Template
}

func NewStartingPointController(
	mux *http.ServeMux,
	planSvc ifaces.PlanService,
	startingPointSvc ifaces.StartingPointService,
	templateCache map[string]*template.Template,
	accessMW *PlanAccessMiddleware,
) *StartingPointController {
	c := &StartingPointController{planSvc: planSvc, startingPointSvc: startingPointSvc, templateCache: templateCache}
	viewer := accessMW.RequireAccess(domain.Viewer)
	editor := accessMW.RequireAccess(domain.Editor)

	mux.Handle("GET /plan/{id}/starting-point", viewer(http.HandlerFunc(c.GetStartingPointSummary)))
	mux.Handle("GET /plan/{id}/starting-point/intro/{section}", viewer(http.HandlerFunc(c.GetStartingPointSectionIntro)))

	mux.Handle("GET /plan/{id}/starting-point/fixed-assets", viewer(http.HandlerFunc(c.GetFixedAssetList)))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/new", editor(http.HandlerFunc(c.PostFixedAssetNew)))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/finish", editor(http.HandlerFunc(c.PostFixedAssetFinish)))
	mux.Handle("GET /plan/{id}/starting-point/fixed-assets/{itemID}/{step}", viewer(http.HandlerFunc(c.GetFixedAssetStep)))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/{itemID}/{step}", editor(http.HandlerFunc(c.PostFixedAssetStep)))
	mux.Handle("POST /plan/{id}/starting-point/fixed-assets/{itemID}", editor(http.HandlerFunc(c.PostFixedAssetDelete)))

	mux.Handle("GET /plan/{id}/starting-point/startup-costs", viewer(http.HandlerFunc(c.GetStartupCostList)))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/new", editor(http.HandlerFunc(c.PostStartupCostNew)))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/finish", editor(http.HandlerFunc(c.PostStartupCostFinish)))
	mux.Handle("GET /plan/{id}/starting-point/startup-costs/{itemID}/{step}", viewer(http.HandlerFunc(c.GetStartupCostStep)))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/{itemID}/{step}", editor(http.HandlerFunc(c.PostStartupCostStep)))
	mux.Handle("POST /plan/{id}/starting-point/startup-costs/{itemID}", editor(http.HandlerFunc(c.PostStartupCostDelete)))

	mux.Handle("GET /plan/{id}/starting-point/funding-sources", viewer(http.HandlerFunc(c.GetFundingSourceList)))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/new", editor(http.HandlerFunc(c.PostFundingSourceNew)))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/finish", editor(http.HandlerFunc(c.PostFundingSourceFinish)))
	mux.Handle("GET /plan/{id}/starting-point/funding-sources/{itemID}/{step}", viewer(http.HandlerFunc(c.GetFundingSourceStep)))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/{itemID}/{step}", editor(http.HandlerFunc(c.PostFundingSourceStep)))
	mux.Handle("POST /plan/{id}/starting-point/funding-sources/{itemID}", editor(http.HandlerFunc(c.PostFundingSourceDelete)))

	mux.Handle("GET /plan/{id}/starting-point/cash-on-hand", viewer(http.HandlerFunc(c.GetCashOnHandEntry)))
	mux.Handle("GET /plan/{id}/starting-point/cash-on-hand/{step}", viewer(http.HandlerFunc(c.GetCashOnHandStep)))
	mux.Handle("POST /plan/{id}/starting-point/cash-on-hand/{step}", editor(http.HandlerFunc(c.PostCashOnHandStep)))

	return c
}

// startingPointComplete reports whether every Starting Point sub-section is
// complete for planID. Used to drive the sidebar's Starting Point nav icon,
// which is shown on every page, not just Starting Point's own.
func (c *StartingPointController) startingPointComplete(planID uuid.UUID) bool {
	status, err := c.startingPointSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
		return false
	}
	return views.IsStartingPointComplete(status)
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
func (c *StartingPointController) GetStartingPointSectionIntro(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	section := r.PathValue("section")
	if !isStartingPointSection(section) {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That Starting Point section doesn't exist.")
		return
	}

	page := views.BuildStartingPointSectionIntroPage(r, user, plan, section, c.startingPointComplete(planID))
	views.RenderSectionIntroPage(w, c.templateCache, page)
}

// GetStartingPointSummary GET /plan/{id}/starting-point
func (c *StartingPointController) GetStartingPointSummary(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "The Plan ID in the URL is malformed or invalid.")
		return
	}

	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that business plan. It may have been deleted, or you might be using an old link.")
		return
	}

	sectionStatus, err := c.startingPointSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
		sectionStatus = map[string]bool{}
	}

	fixedAssets, err := c.startingPointSvc.ListCompleteCapitalAssets(planID)
	if err != nil {
		log.Printf("Failed to load capital assets for plan %s: %v", planID, err)
	}
	startupCosts, err := c.startingPointSvc.ListCompleteStartupCosts(planID)
	if err != nil {
		log.Printf("Failed to load startup costs for plan %s: %v", planID, err)
	}
	fundingSources, err := c.startingPointSvc.ListCompleteFundingSources(planID)
	if err != nil {
		log.Printf("Failed to load funding sources for plan %s: %v", planID, err)
	}

	page := views.BuildStartingPointSummaryPage(r, user, plan, sectionStatus, len(fixedAssets), len(startupCosts), len(fundingSources))
	views.RenderHubSummaryPage(w, c.templateCache, page)
}

// --- Fixed Assets ---

// GetFixedAssetList GET /plan/{id}/starting-point/fixed-assets
func (c *StartingPointController) GetFixedAssetList(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	completeItems, err := c.startingPointSvc.ListCompleteCapitalAssets(planID)
	if err != nil {
		log.Printf("Failed to load capital assets for plan %s: %v", planID, err)
	}
	items := make([]views.StartingPointFixedAssetItem, len(completeItems))
	for i, it := range completeItems {
		items[i] = views.StartingPointFixedAssetItem{ID: it.ID, Asset: it.Asset}
	}

	sectionStatus, err := c.startingPointSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
	}

	var draftItemID *uuid.UUID
	var draftStep string
	if draft, err := c.startingPointSvc.FindCapitalAssetDraft(planID); err != nil {
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
	views.RenderSectionListPage(w, c.templateCache, page)
}

// PostFixedAssetNew POST /plan/{id}/starting-point/fixed-assets/new
// Starts a new Fixed Asset item, or resumes the plan's existing draft if
// one is already in progress.
func (c *StartingPointController) PostFixedAssetNew(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	itemID, err := c.startingPointSvc.CreateCapitalAssetDraft(planID)
	if err != nil {
		log.Printf("Failed to create capital asset draft for plan %s: %v", planID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to start a new fixed asset. Please try again.")
		return
	}

	http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionFixedAssets, fixedAssetSteps[0]), http.StatusSeeOther)
}

// GetFixedAssetStep GET /plan/{id}/starting-point/fixed-assets/{itemID}/{step}
func (c *StartingPointController) GetFixedAssetStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(fixedAssetSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.startingPointSvc.GetCapitalAsset(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}

	backURL := ""
	if prev := prevStepName(fixedAssetSteps, idx); prev != "" {
		backURL = views.SectionStepURL(planID, itemID, domain.SectionFixedAssets, prev)
	}

	page := views.BuildFixedAssetStepPage(r, user, plan, itemID, item.Asset, step, idx+1, len(fixedAssetSteps), backURL, "Next", "", c.startingPointComplete(planID))
	views.RenderFixedAssetStepPage(w, c.templateCache, page)
}

// PostFixedAssetStep POST /plan/{id}/starting-point/fixed-assets/{itemID}/{step}
func (c *StartingPointController) PostFixedAssetStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(fixedAssetSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.startingPointSvc.GetCapitalAsset(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}
	asset := item.Asset
	wasComplete := item.Status == repositories.StatusComplete

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(fixedAssetSteps, idx); prev != "" {
			backURL = views.SectionStepURL(planID, itemID, domain.SectionFixedAssets, prev)
		}
		page := views.BuildFixedAssetStepPage(r, user, plan, itemID, asset, step, idx+1, len(fixedAssetSteps), backURL, "Next", errMsg, c.startingPointComplete(planID))
		views.RenderFixedAssetStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
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
		if err != nil || years <= 0 || years > 50 {
			renderStepError("Please enter a valid useful life in years (1-50).")
			return
		}
		asset.UsefulLifeMonths = years * 12
	}

	// Land doesn't depreciate: a None method finishes the item right
	// after this step, skipping "useful-life" entirely.
	finishNow := step == "useful-life" || (step == "depreciation-method" && asset.DepreciationMethod == domain.None)

	// Full cross-field validation (domain.ValidateCapitalAsset, run by the
	// service on StatusComplete) only makes sense once every field has
	// actually been answered - running it after every intermediate step
	// spuriously rejects a fresh draft (e.g. DepreciationMethod is still
	// "" on the "name"/"cost" steps, tripping the "must have a useful
	// life unless None" rule before the user has even reached that
	// question). finishNow gates when we ask for StatusComplete at all.
	newStatus := repositories.StatusDraft
	newCurrentStep := idx + 1
	if finishNow {
		newStatus = repositories.StatusComplete
		newCurrentStep = len(fixedAssetSteps)
	}

	if err := c.startingPointSvc.SaveCapitalAssetStep(itemID, asset, newCurrentStep, newStatus); err != nil {
		if finishNow {
			renderStepError(err.Error())
		} else {
			log.Printf("Failed to save fixed asset step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
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

	page := views.BuildFixedAssetsAddAnotherPage(r, user, plan, asset, c.startingPointComplete(planID))
	views.RenderAddAnotherPage(w, c.templateCache, page)
}

// PostFixedAssetDelete POST /plan/{id}/starting-point/fixed-assets/{itemID}
func (c *StartingPointController) PostFixedAssetDelete(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	if err := c.startingPointSvc.DeleteCapitalAsset(itemID); err != nil {
		log.Printf("Failed to delete capital asset %s: %v", itemID, err)
	}
	http.Redirect(w, r, views.SectionListURL(planID, domain.SectionFixedAssets), http.StatusSeeOther)
}

// PostFixedAssetFinish POST /plan/{id}/starting-point/fixed-assets/finish
func (c *StartingPointController) PostFixedAssetFinish(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	if err := c.startingPointSvc.MarkWizardSectionComplete(planID, domain.SectionFixedAssets); err != nil {
		log.Printf("Failed to mark fixed assets complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
}

// --- Startup Costs ---

// GetStartupCostList GET /plan/{id}/starting-point/startup-costs
func (c *StartingPointController) GetStartupCostList(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	completeItems, err := c.startingPointSvc.ListCompleteStartupCosts(planID)
	if err != nil {
		log.Printf("Failed to load startup costs for plan %s: %v", planID, err)
	}
	items := make([]views.StartingPointStartupCostItem, len(completeItems))
	for i, it := range completeItems {
		items[i] = views.StartingPointStartupCostItem{ID: it.ID, Cost: it.Cost}
	}

	sectionStatus, err := c.startingPointSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
	}

	var draftItemID *uuid.UUID
	var draftStep string
	if draft, err := c.startingPointSvc.FindStartupCostDraft(planID); err != nil {
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
	views.RenderSectionListPage(w, c.templateCache, page)
}

// PostStartupCostNew POST /plan/{id}/starting-point/startup-costs/new
func (c *StartingPointController) PostStartupCostNew(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	itemID, err := c.startingPointSvc.CreateStartupCostDraft(planID)
	if err != nil {
		log.Printf("Failed to create startup cost draft for plan %s: %v", planID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to start a new startup cost. Please try again.")
		return
	}

	http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionStartupCosts, startupCostSteps[0]), http.StatusSeeOther)
}

// GetStartupCostStep GET /plan/{id}/starting-point/startup-costs/{itemID}/{step}
func (c *StartingPointController) GetStartupCostStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(startupCostSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.startingPointSvc.GetStartupCost(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}

	backURL := ""
	if prev := prevStepName(startupCostSteps, idx); prev != "" {
		backURL = views.SectionStepURL(planID, itemID, domain.SectionStartupCosts, prev)
	}

	page := views.BuildStartupCostStepPage(r, user, plan, itemID, item.Cost, step, idx+1, len(startupCostSteps), backURL, "Next", "", c.startingPointComplete(planID))
	views.RenderStartupCostStepPage(w, c.templateCache, page)
}

// PostStartupCostStep POST /plan/{id}/starting-point/startup-costs/{itemID}/{step}
func (c *StartingPointController) PostStartupCostStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(startupCostSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.startingPointSvc.GetStartupCost(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}
	cost := item.Cost
	wasComplete := item.Status == repositories.StatusComplete

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(startupCostSteps, idx); prev != "" {
			backURL = views.SectionStepURL(planID, itemID, domain.SectionStartupCosts, prev)
		}
		page := views.BuildStartupCostStepPage(r, user, plan, itemID, cost, step, idx+1, len(startupCostSteps), backURL, "Next", errMsg, c.startingPointComplete(planID))
		views.RenderStartupCostStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
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

	// Full validation (domain.ValidateStartupCost, run by the service on
	// StatusComplete) only makes sense once every field has actually been
	// answered - see the matching comment in PostFixedAssetStep.
	newStatus := repositories.StatusDraft
	newCurrentStep := idx + 1
	if finishNow {
		newStatus = repositories.StatusComplete
	}

	if err := c.startingPointSvc.SaveStartupCostStep(itemID, cost, newCurrentStep, newStatus); err != nil {
		if finishNow {
			renderStepError(err.Error())
		} else {
			log.Printf("Failed to save startup cost step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
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

	page := views.BuildStartupCostsAddAnotherPage(r, user, plan, cost, c.startingPointComplete(planID))
	views.RenderAddAnotherPage(w, c.templateCache, page)
}

// PostStartupCostDelete POST /plan/{id}/starting-point/startup-costs/{itemID}
func (c *StartingPointController) PostStartupCostDelete(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	if err := c.startingPointSvc.DeleteStartupCost(itemID); err != nil {
		log.Printf("Failed to delete startup cost %s: %v", itemID, err)
	}
	http.Redirect(w, r, views.SectionListURL(planID, domain.SectionStartupCosts), http.StatusSeeOther)
}

// PostStartupCostFinish POST /plan/{id}/starting-point/startup-costs/finish
func (c *StartingPointController) PostStartupCostFinish(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	if err := c.startingPointSvc.MarkWizardSectionComplete(planID, domain.SectionStartupCosts); err != nil {
		log.Printf("Failed to mark startup costs complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
}

// --- Funding Sources ---

// GetFundingSourceList GET /plan/{id}/starting-point/funding-sources
func (c *StartingPointController) GetFundingSourceList(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}

	completeItems, err := c.startingPointSvc.ListCompleteFundingSources(planID)
	if err != nil {
		log.Printf("Failed to load funding sources for plan %s: %v", planID, err)
	}
	items := make([]views.StartingPointFundingSourceItem, len(completeItems))
	for i, it := range completeItems {
		items[i] = views.StartingPointFundingSourceItem{ID: it.ID, Funding: it.Funding}
	}

	sectionStatus, err := c.startingPointSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
	}

	var draftItemID *uuid.UUID
	var draftStep string
	if draft, err := c.startingPointSvc.FindFundingSourceDraft(planID); err != nil {
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
	views.RenderSectionListPage(w, c.templateCache, page)
}

// PostFundingSourceNew POST /plan/{id}/starting-point/funding-sources/new
func (c *StartingPointController) PostFundingSourceNew(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	itemID, err := c.startingPointSvc.CreateFundingSourceDraft(planID)
	if err != nil {
		log.Printf("Failed to create funding source draft for plan %s: %v", planID, err)
		renderErrorPage(w, r, c.templateCache, http.StatusInternalServerError, "Failed to start a new funding source. Please try again.")
		return
	}

	http.Redirect(w, r, views.SectionStepURL(planID, itemID, domain.SectionFundingSources, fundingSourceSteps[0]), http.StatusSeeOther)
}

// GetFundingSourceStep GET /plan/{id}/starting-point/funding-sources/{itemID}/{step}
func (c *StartingPointController) GetFundingSourceStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(fundingSourceSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.startingPointSvc.GetFundingSource(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}

	backURL := ""
	if prev := prevStepName(fundingSourceSteps, idx); prev != "" {
		backURL = views.SectionStepURL(planID, itemID, domain.SectionFundingSources, prev)
	}

	page := views.BuildFundingSourceStepPage(r, user, plan, itemID, item.Funding, step, idx+1, len(fundingSourceSteps), backURL, "Next", "", c.startingPointComplete(planID))
	views.RenderFundingSourceStepPage(w, c.templateCache, page)
}

// PostFundingSourceStep POST /plan/{id}/starting-point/funding-sources/{itemID}/{step}
func (c *StartingPointController) PostFundingSourceStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(fundingSourceSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	item, err := c.startingPointSvc.GetFundingSource(itemID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "We couldn't find that item. It may have been deleted.")
		return
	}
	funding := item.Funding
	wasComplete := item.Status == repositories.StatusComplete

	renderStepError := func(errMsg string) {
		backURL := ""
		if prev := prevStepName(fundingSourceSteps, idx); prev != "" {
			backURL = views.SectionStepURL(planID, itemID, domain.SectionFundingSources, prev)
		}
		page := views.BuildFundingSourceStepPage(r, user, plan, itemID, funding, step, idx+1, len(fundingSourceSteps), backURL, "Next", errMsg, c.startingPointComplete(planID))
		views.RenderFundingSourceStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
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

	// Full validation (domain.ValidateFundingSource, run by the service on
	// StatusComplete) only makes sense once every field has actually been
	// answered - see the matching comment in PostFixedAssetStep.
	newStatus := repositories.StatusDraft
	newCurrentStep := idx + 1
	if finishNow {
		newStatus = repositories.StatusComplete
	}

	if err := c.startingPointSvc.SaveFundingSourceStep(itemID, funding, newCurrentStep, newStatus); err != nil {
		if finishNow {
			renderStepError(err.Error())
		} else {
			log.Printf("Failed to save funding source step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
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

	page := views.BuildFundingSourcesAddAnotherPage(r, user, plan, funding, c.startingPointComplete(planID))
	views.RenderAddAnotherPage(w, c.templateCache, page)
}

// PostFundingSourceDelete POST /plan/{id}/starting-point/funding-sources/{itemID}
func (c *StartingPointController) PostFundingSourceDelete(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemID"))
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid item ID")
		return
	}
	if err := c.startingPointSvc.DeleteFundingSource(itemID); err != nil {
		log.Printf("Failed to delete funding source %s: %v", itemID, err)
	}
	http.Redirect(w, r, views.SectionListURL(planID, domain.SectionFundingSources), http.StatusSeeOther)
}

// PostFundingSourceFinish POST /plan/{id}/starting-point/funding-sources/finish
func (c *StartingPointController) PostFundingSourceFinish(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	if err := c.startingPointSvc.MarkWizardSectionComplete(planID, domain.SectionFundingSources); err != nil {
		log.Printf("Failed to mark funding sources complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
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
func (c *StartingPointController) GetCashOnHandEntry(w http.ResponseWriter, r *http.Request) {
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	sectionStatus, err := c.startingPointSvc.GetHubStatus(planID)
	if err != nil {
		log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
	}

	step := cashOnHandSteps[0]
	if !sectionStatus[domain.SectionCashOnHand] {
		if row, err := c.startingPointSvc.GetStartingBalancesRow(planID); err != nil {
			log.Printf("Failed to load starting balances for plan %s: %v", planID, err)
		} else if row.CurrentStep > 0 && row.CurrentStep < len(cashOnHandSteps) {
			step = cashOnHandSteps[row.CurrentStep]
		}
	}

	http.Redirect(w, r, views.SectionSingletonStepURL(planID, domain.SectionCashOnHand, step), http.StatusSeeOther)
}

// GetCashOnHandStep GET /plan/{id}/starting-point/cash-on-hand/{step}
func (c *StartingPointController) GetCashOnHandStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(cashOnHandSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	row, err := c.startingPointSvc.GetStartingBalancesRow(planID)
	if err != nil {
		log.Printf("Failed to load starting balances for plan %s: %v", planID, err)
		row = &repositories.StartingBalancesRow{}
	}

	backURL := ""
	if prev := prevStepName(cashOnHandSteps, idx); prev != "" {
		backURL = views.SectionSingletonStepURL(planID, domain.SectionCashOnHand, prev)
	}

	page := views.BuildCashOnHandStepPage(r, user, plan, row.Balances, step, idx+1, len(cashOnHandSteps), backURL, cashOnHandButtonLabel(idx), "", c.startingPointComplete(planID))
	views.RenderCashOnHandStepPage(w, c.templateCache, page)
}

// PostCashOnHandStep POST /plan/{id}/starting-point/cash-on-hand/{step}
func (c *StartingPointController) PostCashOnHandStep(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	planID, err := parsePlanID(r)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "Invalid plan ID")
		return
	}
	plan, err := c.planSvc.Get(planID)
	if err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "Plan not found")
		return
	}
	step := r.PathValue("step")
	idx := stepIndex(cashOnHandSteps, step)
	if idx == -1 {
		renderErrorPage(w, r, c.templateCache, http.StatusNotFound, "That step doesn't exist.")
		return
	}

	row, err := c.startingPointSvc.GetStartingBalancesRow(planID)
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
		page := views.BuildCashOnHandStepPage(r, user, plan, balances, step, idx+1, len(cashOnHandSteps), backURL, cashOnHandButtonLabel(idx), errMsg, c.startingPointComplete(planID))
		views.RenderCashOnHandStepPageWithStatus(w, c.templateCache, page, http.StatusBadRequest)
	}

	amount, ok := parseStepMoney(r.PostForm.Get("amount"))
	if !ok {
		renderStepError("Please enter a valid amount.")
		return
	}
	balances = cashOnHandFieldSet(balances, step, amount)

	newCurrentStep := idx + 1
	isLastStep := idx == len(cashOnHandSteps)-1
	if err := c.startingPointSvc.SaveStartingBalancesStep(planID, balances, newCurrentStep, isLastStep); err != nil {
		if isLastStep {
			renderStepError(err.Error())
		} else {
			log.Printf("Failed to save starting balances step: %v", err)
			renderStepError("An internal database error occurred. Please try again.")
		}
		return
	}

	if idx < len(cashOnHandSteps)-1 {
		http.Redirect(w, r, views.SectionSingletonStepURL(planID, domain.SectionCashOnHand, nextStepName(cashOnHandSteps, idx)), http.StatusSeeOther)
		return
	}

	// Last step: Cash on Hand isn't repeatable, so finishing it marks
	// the section complete directly and returns to the summary page.
	if err := c.startingPointSvc.MarkWizardSectionComplete(planID, domain.SectionCashOnHand); err != nil {
		log.Printf("Failed to mark cash on hand complete for plan %s: %v", planID, err)
	}
	http.Redirect(w, r, views.StartingPointSummaryURL(planID), http.StatusSeeOther)
}
