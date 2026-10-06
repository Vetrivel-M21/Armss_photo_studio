package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/store/mysql"
)

type UsageHandler struct {
	limitsRepo  *mysql.UserLimitsRepo
	recordsRepo *mysql.RateRecordsRepo
}

func NewUsageHandler(limitsRepo *mysql.UserLimitsRepo, recordsRepo *mysql.RateRecordsRepo) *UsageHandler {
	return &UsageHandler{limitsRepo: limitsRepo, recordsRepo: recordsRepo}
}

type UserLimitDTO struct {
	MaxAPICallsPerHour  int  `json:"maxApiCallsPerHour"`
	MaxRendersPerHour   int  `json:"maxRendersPerHour"`
	MonthlySpendCapCents int `json:"monthlySpendCapCents"`
	IsUnlimited         bool `json:"isUnlimited"`
	IsCustom            bool `json:"isCustom"`
}

type UserUsageDTO struct {
	UserID         string       `json:"userId"`
	Email          string       `json:"email"`
	Role           string       `json:"role"`
	CreatedAt      string       `json:"createdAt"`
	TotalProjects  int          `json:"totalProjects"`
	TotalTokensIn  int          `json:"totalTokensIn"`
	TotalTokensOut int          `json:"totalTokensOut"`
	TotalRenders   int          `json:"totalRenders"`
	TotalCostCents int          `json:"totalCostCents"`
	HourAPICalls   int          `json:"hourApiCalls"`
	HourRenders    int          `json:"hourRenders"`
	Limits         UserLimitDTO `json:"limits"`
}

type UsageOverviewResponse struct {
	TotalProjects  int            `json:"totalProjects"`
	TotalTokensIn  int            `json:"totalTokensIn"`
	TotalTokensOut int            `json:"totalTokensOut"`
	TotalRenders   int            `json:"totalRenders"`
	TotalCostCents int            `json:"totalCostCents"`
	GlobalLimit    UserLimitDTO   `json:"globalLimit"`
	Users          []UserUsageDTO `json:"users"`
}

// GetUsage implements GET /api/v1/admin/usage.
func (h *UsageHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	globalLimit, err := h.limitsRepo.GetGlobalDefault(ctx)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load global limits.")
		return
	}

	rows, err := h.recordsRepo.GetUserUsageOverview(ctx)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load usage overview.")
		return
	}

	customLimits, err := h.limitsRepo.ListAll(ctx)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load user limits.")
		return
	}

	customLimitMap := make(map[string]mysql.UserLimit, len(customLimits))
	for _, cl := range customLimits {
		if cl.UserID.Valid {
			customLimitMap[cl.UserID.String] = cl
		}
	}

	resp := UsageOverviewResponse{
		GlobalLimit: UserLimitDTO{
			MaxAPICallsPerHour:  globalLimit.MaxAPICallsPerHour,
			MaxRendersPerHour:   globalLimit.MaxRendersPerHour,
			MonthlySpendCapCents: globalLimit.MonthlySpendCapCents,
			IsUnlimited:         globalLimit.IsUnlimited,
			IsCustom:            false,
		},
		Users: make([]UserUsageDTO, 0, len(rows)),
	}

	for _, row := range rows {
		tokensIn := 0
		if row.TotalTokensIn.Valid {
			tokensIn = int(row.TotalTokensIn.Int64)
		}
		tokensOut := 0
		if row.TotalTokensOut.Valid {
			tokensOut = int(row.TotalTokensOut.Int64)
		}
		costCents := 0
		if row.TotalCostCents.Valid {
			costCents = int(row.TotalCostCents.Int64)
		}

		hourCalls, _ := h.recordsRepo.CountInPastHour(ctx, row.UserID, mysql.ActionPipeline)
		hourRenders, _ := h.recordsRepo.CountInPastHour(ctx, row.UserID, mysql.ActionRender)

		limitDTO := resp.GlobalLimit
		if cl, exists := customLimitMap[row.UserID]; exists {
			limitDTO = UserLimitDTO{
				MaxAPICallsPerHour:  cl.MaxAPICallsPerHour,
				MaxRendersPerHour:   cl.MaxRendersPerHour,
				MonthlySpendCapCents: cl.MonthlySpendCapCents,
				IsUnlimited:         cl.IsUnlimited,
				IsCustom:            true,
			}
		}

		userDTO := UserUsageDTO{
			UserID:         row.UserID,
			Email:          row.Email,
			Role:           row.Role,
			CreatedAt:      row.CreatedAt.Format("2006-01-02T15:04:05Z"),
			TotalProjects:  row.TotalProjects,
			TotalTokensIn:  tokensIn,
			TotalTokensOut: tokensOut,
			TotalRenders:   row.TotalRenders,
			TotalCostCents: costCents,
			HourAPICalls:   hourCalls,
			HourRenders:    hourRenders,
			Limits:         limitDTO,
		}

		resp.TotalProjects += row.TotalProjects
		resp.TotalTokensIn += tokensIn
		resp.TotalTokensOut += tokensOut
		resp.TotalRenders += row.TotalRenders
		resp.TotalCostCents += costCents

		resp.Users = append(resp.Users, userDTO)
	}

	respond.JSON(w, http.StatusOK, resp)
}

type UpdateLimitsRequest struct {
	MaxAPICallsPerHour  int  `json:"maxApiCallsPerHour"`
	MaxRendersPerHour   int  `json:"maxRendersPerHour"`
	MonthlySpendCapCents int `json:"monthlySpendCapCents"`
	IsUnlimited         bool `json:"isUnlimited,omitempty"`
}

// UpdateGlobalLimits implements PUT /api/v1/admin/limits.
func (h *UsageHandler) UpdateGlobalLimits(w http.ResponseWriter, r *http.Request) {
	var req UpdateLimitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	if req.MaxAPICallsPerHour <= 0 || req.MaxRendersPerHour <= 0 {
		respond.Error(w, http.StatusBadRequest, "Hourly limits must be positive integers.")
		return
	}

	if err := h.limitsRepo.SetGlobalDefault(r.Context(), req.MaxAPICallsPerHour, req.MaxRendersPerHour, req.MonthlySpendCapCents); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not update global limits.")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"message": "Global limits updated successfully."})
}

// UpdateUserLimits implements PUT /api/v1/admin/users/{id}/limits.
func (h *UsageHandler) UpdateUserLimits(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		respond.Error(w, http.StatusBadRequest, "User ID is required.")
		return
	}

	var req UpdateLimitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	if !req.IsUnlimited && (req.MaxAPICallsPerHour <= 0 || req.MaxRendersPerHour <= 0) {
		respond.Error(w, http.StatusBadRequest, "Hourly limits must be positive integers.")
		return
	}

	if err := h.limitsRepo.SetForUser(r.Context(), userID, req.MaxAPICallsPerHour, req.MaxRendersPerHour, req.MonthlySpendCapCents, req.IsUnlimited); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not update user limits.")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"message": "User limits updated successfully."})
}
