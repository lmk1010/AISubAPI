package admin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ChannelHandler handles admin channel management
type ChannelHandler struct {
	channelService *service.ChannelService
	billingService *service.BillingService
	pricingService *service.PricingService
}

// NewChannelHandler creates a new admin channel handler
func NewChannelHandler(channelService *service.ChannelService, billingService *service.BillingService, pricingService *service.PricingService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
		billingService: billingService,
		pricingService: pricingService,
	}
}

// --- Request / Response types ---

type createChannelRequest struct {
	Name                       string                           `json:"name" binding:"required,max=100"`
	Description                string                           `json:"description"`
	GroupIDs                   []int64                          `json:"group_ids"`
	ModelPricing               []channelModelPricingRequest     `json:"model_pricing"`
	ModelMapping               map[string]map[string]string     `json:"model_mapping"`
	BillingModelSource         string                           `json:"billing_model_source" binding:"omitempty,oneof=requested upstream channel_mapped"`
	RestrictModels             bool                             `json:"restrict_models"`
	Features                   string                           `json:"features"`
	FeaturesConfig             map[string]any                   `json:"features_config"`
	ApplyPricingToAccountStats bool                             `json:"apply_pricing_to_account_stats"`
	AccountStatsPricingRules   []accountStatsPricingRuleRequest `json:"account_stats_pricing_rules"`
}

type updateChannelRequest struct {
	Name                       string                            `json:"name" binding:"omitempty,max=100"`
	Description                *string                           `json:"description"`
	Status                     string                            `json:"status" binding:"omitempty,oneof=active disabled"`
	GroupIDs                   *[]int64                          `json:"group_ids"`
	ModelPricing               *[]channelModelPricingRequest     `json:"model_pricing"`
	ModelMapping               map[string]map[string]string      `json:"model_mapping"`
	BillingModelSource         string                            `json:"billing_model_source" binding:"omitempty,oneof=requested upstream channel_mapped"`
	RestrictModels             *bool                             `json:"restrict_models"`
	Features                   *string                           `json:"features"`
	FeaturesConfig             map[string]any                    `json:"features_config"`
	ApplyPricingToAccountStats *bool                             `json:"apply_pricing_to_account_stats"`
	AccountStatsPricingRules   *[]accountStatsPricingRuleRequest `json:"account_stats_pricing_rules"`
}

type channelModelPricingRequest struct {
	Platform         string                   `json:"platform" binding:"omitempty,max=50"`
	Models           []string                 `json:"models" binding:"required,min=1,max=100"`
	BillingMode      string                   `json:"billing_mode" binding:"omitempty,oneof=token per_request image"`
	InputPrice       *float64                 `json:"input_price" binding:"omitempty,min=0"`
	OutputPrice      *float64                 `json:"output_price" binding:"omitempty,min=0"`
	CacheWritePrice  *float64                 `json:"cache_write_price" binding:"omitempty,min=0"`
	CacheReadPrice   *float64                 `json:"cache_read_price" binding:"omitempty,min=0"`
	ImageOutputPrice *float64                 `json:"image_output_price" binding:"omitempty,min=0"`
	PerRequestPrice  *float64                 `json:"per_request_price" binding:"omitempty,min=0"`
	Intervals        []pricingIntervalRequest `json:"intervals"`
}

type pricingIntervalRequest struct {
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
	SortOrder       int      `json:"sort_order"`
}

type accountStatsPricingRuleRequest struct {
	Name       string                       `json:"name"`
	GroupIDs   []int64                      `json:"group_ids"`
	AccountIDs []int64                      `json:"account_ids"`
	Pricing    []channelModelPricingRequest `json:"pricing"`
}

type channelResponse struct {
	ID                         int64                             `json:"id"`
	Name                       string                            `json:"name"`
	Description                string                            `json:"description"`
	Status                     string                            `json:"status"`
	BillingModelSource         string                            `json:"billing_model_source"`
	RestrictModels             bool                              `json:"restrict_models"`
	Features                   string                            `json:"features"`
	FeaturesConfig             map[string]any                    `json:"features_config"`
	GroupIDs                   []int64                           `json:"group_ids"`
	ModelPricing               []channelModelPricingResponse     `json:"model_pricing"`
	ModelMapping               map[string]map[string]string      `json:"model_mapping"`
	ApplyPricingToAccountStats bool                              `json:"apply_pricing_to_account_stats"`
	AccountStatsPricingRules   []accountStatsPricingRuleResponse `json:"account_stats_pricing_rules"`
	CreatedAt                  string                            `json:"created_at"`
	UpdatedAt                  string                            `json:"updated_at"`
}

type channelModelPricingResponse struct {
	ID               int64                     `json:"id"`
	Platform         string                    `json:"platform"`
	Models           []string                  `json:"models"`
	BillingMode      string                    `json:"billing_mode"`
	InputPrice       *float64                  `json:"input_price"`
	OutputPrice      *float64                  `json:"output_price"`
	CacheWritePrice  *float64                  `json:"cache_write_price"`
	CacheReadPrice   *float64                  `json:"cache_read_price"`
	ImageOutputPrice *float64                  `json:"image_output_price"`
	PerRequestPrice  *float64                  `json:"per_request_price"`
	Intervals        []pricingIntervalResponse `json:"intervals"`
}

type pricingIntervalResponse struct {
	ID              int64    `json:"id"`
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label,omitempty"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
	SortOrder       int      `json:"sort_order"`
}

type accountStatsPricingRuleResponse struct {
	ID         int64                         `json:"id"`
	Name       string                        `json:"name"`
	GroupIDs   []int64                       `json:"group_ids"`
	AccountIDs []int64                       `json:"account_ids"`
	Pricing    []channelModelPricingResponse `json:"pricing"`
}

func channelToResponse(ch *service.Channel) *channelResponse {
	if ch == nil {
		return nil
	}
	resp := &channelResponse{
		ID:             ch.ID,
		Name:           ch.Name,
		Description:    ch.Description,
		Status:         ch.Status,
		RestrictModels: ch.RestrictModels,
		Features:       ch.Features,
		FeaturesConfig: ch.FeaturesConfig,
		GroupIDs:       ch.GroupIDs,
		ModelMapping:   ch.ModelMapping,
		CreatedAt:      ch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      ch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp.BillingModelSource = ch.BillingModelSource
	if resp.GroupIDs == nil {
		resp.GroupIDs = []int64{}
	}
	if resp.ModelMapping == nil {
		resp.ModelMapping = map[string]map[string]string{}
	}

	resp.ModelPricing = make([]channelModelPricingResponse, 0, len(ch.ModelPricing))
	for _, p := range ch.ModelPricing {
		resp.ModelPricing = append(resp.ModelPricing, pricingToResponse(&p))
	}

	resp.ApplyPricingToAccountStats = ch.ApplyPricingToAccountStats
	resp.AccountStatsPricingRules = make([]accountStatsPricingRuleResponse, 0, len(ch.AccountStatsPricingRules))
	for _, rule := range ch.AccountStatsPricingRules {
		ruleResp := accountStatsPricingRuleResponse{
			ID:         rule.ID,
			Name:       rule.Name,
			GroupIDs:   rule.GroupIDs,
			AccountIDs: rule.AccountIDs,
			Pricing:    make([]channelModelPricingResponse, 0, len(rule.Pricing)),
		}
		if ruleResp.GroupIDs == nil {
			ruleResp.GroupIDs = []int64{}
		}
		if ruleResp.AccountIDs == nil {
			ruleResp.AccountIDs = []int64{}
		}
		for i := range rule.Pricing {
			ruleResp.Pricing = append(ruleResp.Pricing, pricingToResponse(&rule.Pricing[i]))
		}
		resp.AccountStatsPricingRules = append(resp.AccountStatsPricingRules, ruleResp)
	}

	return resp
}

func pricingToResponse(p *service.ChannelModelPricing) channelModelPricingResponse {
	models := p.Models
	if models == nil {
		models = []string{}
	}
	billingMode := string(p.BillingMode)
	if billingMode == "" {
		billingMode = string(service.BillingModeToken)
	}
	platform := p.Platform
	if platform == "" {
		platform = service.PlatformAnthropic
	}
	intervals := make([]pricingIntervalResponse, 0, len(p.Intervals))
	for _, iv := range p.Intervals {
		intervals = append(intervals, intervalToResponse(iv))
	}
	return channelModelPricingResponse{
		ID:               p.ID,
		Platform:         platform,
		Models:           models,
		BillingMode:      billingMode,
		InputPrice:       p.InputPrice,
		OutputPrice:      p.OutputPrice,
		CacheWritePrice:  p.CacheWritePrice,
		CacheReadPrice:   p.CacheReadPrice,
		ImageOutputPrice: p.ImageOutputPrice,
		PerRequestPrice:  p.PerRequestPrice,
		Intervals:        intervals,
	}
}

func intervalToResponse(iv service.PricingInterval) pricingIntervalResponse {
	return pricingIntervalResponse{
		ID:              iv.ID,
		MinTokens:       iv.MinTokens,
		MaxTokens:       iv.MaxTokens,
		TierLabel:       iv.TierLabel,
		InputPrice:      iv.InputPrice,
		OutputPrice:     iv.OutputPrice,
		CacheWritePrice: iv.CacheWritePrice,
		CacheReadPrice:  iv.CacheReadPrice,
		PerRequestPrice: iv.PerRequestPrice,
		SortOrder:       iv.SortOrder,
	}
}

func pricingRequestToService(reqs []channelModelPricingRequest) []service.ChannelModelPricing {
	result := make([]service.ChannelModelPricing, 0, len(reqs))
	for _, r := range reqs {
		billingMode := service.BillingMode(r.BillingMode)
		if billingMode == "" {
			billingMode = service.BillingModeToken
		}
		platform := r.Platform
		intervals := make([]service.PricingInterval, 0, len(r.Intervals))
		for _, iv := range r.Intervals {
			intervals = append(intervals, service.PricingInterval{
				MinTokens:       iv.MinTokens,
				MaxTokens:       iv.MaxTokens,
				TierLabel:       iv.TierLabel,
				InputPrice:      iv.InputPrice,
				OutputPrice:     iv.OutputPrice,
				CacheWritePrice: iv.CacheWritePrice,
				CacheReadPrice:  iv.CacheReadPrice,
				PerRequestPrice: iv.PerRequestPrice,
				SortOrder:       iv.SortOrder,
			})
		}
		result = append(result, service.ChannelModelPricing{
			Platform:         platform,
			Models:           r.Models,
			BillingMode:      billingMode,
			InputPrice:       r.InputPrice,
			OutputPrice:      r.OutputPrice,
			CacheWritePrice:  r.CacheWritePrice,
			CacheReadPrice:   r.CacheReadPrice,
			ImageOutputPrice: r.ImageOutputPrice,
			PerRequestPrice:  r.PerRequestPrice,
			Intervals:        intervals,
		})
	}
	return result
}

func accountStatsPricingRuleRequestToService(r accountStatsPricingRuleRequest) service.AccountStatsPricingRule {
	return service.AccountStatsPricingRule{
		Name:       r.Name,
		GroupIDs:   r.GroupIDs,
		AccountIDs: r.AccountIDs,
		Pricing:    pricingRequestToService(r.Pricing),
	}
}

// --- Handlers ---

// List handles listing channels with pagination
// GET /api/v1/admin/channels
func (h *ChannelHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}

	channels, pag, err := h.channelService.List(c.Request.Context(), pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}, status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*channelResponse, 0, len(channels))
	for i := range channels {
		out = append(out, channelToResponse(&channels[i]))
	}
	response.Paginated(c, out, pag.Total, page, pageSize)
}

// GetByID handles getting a channel by ID
// GET /api/v1/admin/channels/:id
func (h *ChannelHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_CHANNEL_ID", "Invalid channel ID"))
		return
	}

	channel, err := h.channelService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, channelToResponse(channel))
}

// Create handles creating a new channel
// POST /api/v1/admin/channels
func (h *ChannelHandler) Create(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}

	pricing := pricingRequestToService(req.ModelPricing)
	// Main model_pricing requires a platform; default to anthropic for backward compatibility.
	for i := range pricing {
		if pricing[i].Platform == "" {
			pricing[i].Platform = service.PlatformAnthropic
		}
	}

	var statsRules []service.AccountStatsPricingRule
	for i, r := range req.AccountStatsPricingRules {
		if len(r.GroupIDs) == 0 && len(r.AccountIDs) == 0 {
			response.ErrorFrom(c, infraerrors.BadRequest("PRICING_RULE_EMPTY_SCOPE",
				fmt.Sprintf("pricing rule #%d must have at least one group or account", i+1)))
			return
		}
		if len(r.Pricing) == 0 {
			response.ErrorFrom(c, infraerrors.BadRequest("PRICING_RULE_EMPTY_PRICING",
				fmt.Sprintf("pricing rule #%d must have at least one pricing entry", i+1)))
			return
		}
		rule := accountStatsPricingRuleRequestToService(r)
		rule.SortOrder = i
		statsRules = append(statsRules, rule)
	}

	channel, err := h.channelService.Create(c.Request.Context(), &service.CreateChannelInput{
		Name:                       req.Name,
		Description:                req.Description,
		GroupIDs:                   req.GroupIDs,
		ModelPricing:               pricing,
		ModelMapping:               req.ModelMapping,
		BillingModelSource:         req.BillingModelSource,
		RestrictModels:             req.RestrictModels,
		Features:                   req.Features,
		FeaturesConfig:             req.FeaturesConfig,
		ApplyPricingToAccountStats: req.ApplyPricingToAccountStats,
		AccountStatsPricingRules:   statsRules,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, channelToResponse(channel))
}

// Update handles updating a channel
// PUT /api/v1/admin/channels/:id
func (h *ChannelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_CHANNEL_ID", "Invalid channel ID"))
		return
	}

	var req updateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}

	input := &service.UpdateChannelInput{
		Name:                       req.Name,
		Description:                req.Description,
		Status:                     req.Status,
		GroupIDs:                   req.GroupIDs,
		ModelMapping:               req.ModelMapping,
		BillingModelSource:         req.BillingModelSource,
		RestrictModels:             req.RestrictModels,
		Features:                   req.Features,
		FeaturesConfig:             req.FeaturesConfig,
		ApplyPricingToAccountStats: req.ApplyPricingToAccountStats,
	}
	if req.ModelPricing != nil {
		pricing := pricingRequestToService(*req.ModelPricing)
		for i := range pricing {
			if pricing[i].Platform == "" {
				pricing[i].Platform = service.PlatformAnthropic
			}
		}
		input.ModelPricing = &pricing
	}
	if req.AccountStatsPricingRules != nil {
		statsRules := make([]service.AccountStatsPricingRule, 0, len(*req.AccountStatsPricingRules))
		for i, r := range *req.AccountStatsPricingRules {
			if len(r.GroupIDs) == 0 && len(r.AccountIDs) == 0 {
				response.ErrorFrom(c, infraerrors.BadRequest("PRICING_RULE_EMPTY_SCOPE",
					fmt.Sprintf("pricing rule #%d must have at least one group or account", i+1)))
				return
			}
			if len(r.Pricing) == 0 {
				response.ErrorFrom(c, infraerrors.BadRequest("PRICING_RULE_EMPTY_PRICING",
					fmt.Sprintf("pricing rule #%d must have at least one pricing entry", i+1)))
				return
			}
			rule := accountStatsPricingRuleRequestToService(r)
			rule.SortOrder = i
			statsRules = append(statsRules, rule)
		}
		input.AccountStatsPricingRules = &statsRules
	}

	channel, err := h.channelService.Update(c.Request.Context(), id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, channelToResponse(channel))
}

// Delete handles deleting a channel
// DELETE /api/v1/admin/channels/:id
func (h *ChannelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_CHANNEL_ID", "Invalid channel ID"))
		return
	}

	if err := h.channelService.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Channel deleted successfully"})
}

// GetModelDefaultPricing 获取模型的默认定价（用于前端自动填充）
// GET /api/v1/admin/channels/model-pricing?model=claude-sonnet-4
func (h *ChannelHandler) GetModelDefaultPricing(c *gin.Context) {
	model := strings.TrimSpace(c.Query("model"))
	if model == "" {
		response.ErrorFrom(c, infraerrors.BadRequest("MISSING_PARAMETER", "model parameter is required").
			WithMetadata(map[string]string{"param": "model"}))
		return
	}

	pricing, err := h.billingService.GetModelPricing(model)
	if err != nil {
		// 模型不在定价列表中
		response.Success(c, gin.H{"found": false})
		return
	}

	response.Success(c, gin.H{
		"found":              true,
		"input_price":        pricing.InputPricePerToken,
		"output_price":       pricing.OutputPricePerToken,
		"cache_write_price":  pricing.CacheCreationPricePerToken,
		"cache_read_price":   pricing.CacheReadPricePerToken,
		"image_output_price": pricing.ImageOutputPricePerToken,
	})
}

// litellmSuggestionItem is one row of the suggestions response.
type litellmSuggestionItem struct {
	Model             string  `json:"model"`
	InputPrice        float64 `json:"input_price"`
	OutputPrice       float64 `json:"output_price"`
	CacheReadPrice    float64 `json:"cache_read_price"`
	CacheWritePrice   float64 `json:"cache_write_price"`
	ImageOutputPrice  float64 `json:"image_output_price"`
	Mode              string  `json:"mode"`
	SupportsCache     bool    `json:"supports_cache"`
	AlreadyConfigured bool    `json:"already_configured"`
}

// GetLiteLLMSuggestions lists LiteLLM-known models for a given platform,
// flagging those already configured on the channel so the UI can show them
// disabled or hidden.
// GET /api/v1/admin/channels/:id/litellm-suggestions?platform=openai
func (h *ChannelHandler) GetLiteLLMSuggestions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_ID", "invalid channel id"))
		return
	}
	platform := strings.ToLower(strings.TrimSpace(c.Query("platform")))
	if platform == "" {
		response.ErrorFrom(c, infraerrors.BadRequest("MISSING_PARAMETER", "platform is required"))
		return
	}

	channel, err := h.channelService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	existing := make(map[string]struct{})
	for _, p := range channel.ModelPricing {
		if !strings.EqualFold(p.Platform, platform) {
			continue
		}
		for _, m := range p.Models {
			existing[strings.ToLower(strings.TrimSpace(m))] = struct{}{}
		}
	}

	items := h.pricingService.ListByProvider(platform)
	out := make([]litellmSuggestionItem, 0, len(items))
	for name, p := range items {
		_, dup := existing[strings.ToLower(name)]
		out = append(out, litellmSuggestionItem{
			Model:             name,
			InputPrice:        p.InputCostPerToken,
			OutputPrice:       p.OutputCostPerToken,
			CacheReadPrice:    p.CacheReadInputTokenCost,
			CacheWritePrice:   p.CacheCreationInputTokenCost,
			ImageOutputPrice:  p.OutputCostPerImage,
			Mode:              p.Mode,
			SupportsCache:     p.SupportsPromptCaching,
			AlreadyConfigured: dup,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Model < out[j].Model })

	response.Success(c, gin.H{"platform": platform, "items": out})
}

// importLiteLLMRequest is the body for ImportLiteLLMModels.
type importLiteLLMRequest struct {
	Platform string   `json:"platform" binding:"required"`
	Models   []string `json:"models" binding:"required,min=1"`
}

// ImportLiteLLMModels writes LiteLLM official prices into channel_model_pricing
// for the requested model names. Models already present on the channel are
// skipped (operator should manually delete + re-import to refresh prices).
// POST /api/v1/admin/channels/:id/import-litellm-models
func (h *ChannelHandler) ImportLiteLLMModels(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_ID", "invalid channel id"))
		return
	}

	var req importLiteLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_REQUEST", err.Error()))
		return
	}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	if platform == "" {
		response.ErrorFrom(c, infraerrors.BadRequest("MISSING_PARAMETER", "platform is required"))
		return
	}

	channel, err := h.channelService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	existing := make(map[string]struct{})
	for _, p := range channel.ModelPricing {
		if !strings.EqualFold(p.Platform, platform) {
			continue
		}
		for _, m := range p.Models {
			existing[strings.ToLower(strings.TrimSpace(m))] = struct{}{}
		}
	}

	imported := []string{}
	skipped := []string{}
	missing := []string{}
	pricing := append([]service.ChannelModelPricing(nil), channel.ModelPricing...)

	for _, raw := range req.Models {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if _, dup := existing[strings.ToLower(name)]; dup {
			skipped = append(skipped, name)
			continue
		}
		litellm := h.pricingService.GetModelPricing(name)
		if litellm == nil || litellm.InputCostPerToken == 0 {
			missing = append(missing, name)
			continue
		}

		in := litellm.InputCostPerToken
		out := litellm.OutputCostPerToken
		entry := service.ChannelModelPricing{
			ChannelID:   id,
			Platform:    platform,
			Models:      []string{name},
			BillingMode: service.BillingModeToken,
			InputPrice:  &in,
			OutputPrice: &out,
		}
		if litellm.CacheReadInputTokenCost > 0 {
			v := litellm.CacheReadInputTokenCost
			entry.CacheReadPrice = &v
		}
		if litellm.CacheCreationInputTokenCost > 0 {
			v := litellm.CacheCreationInputTokenCost
			entry.CacheWritePrice = &v
		}
		if litellm.OutputCostPerImage > 0 {
			v := litellm.OutputCostPerImage
			entry.ImageOutputPrice = &v
			if litellm.Mode == "image_generation" {
				entry.BillingMode = service.BillingModeImage
			}
		}
		pricing = append(pricing, entry)
		existing[strings.ToLower(name)] = struct{}{}
		imported = append(imported, name)
	}

	if len(imported) == 0 {
		response.Success(c, gin.H{
			"imported": imported,
			"skipped":  skipped,
			"missing":  missing,
		})
		return
	}

	if _, err := h.channelService.Update(c.Request.Context(), id, &service.UpdateChannelInput{
		ModelPricing: &pricing,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"imported": imported,
		"skipped":  skipped,
		"missing":  missing,
	})
}
