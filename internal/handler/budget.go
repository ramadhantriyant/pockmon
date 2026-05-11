package handler

import (
	"errors"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ramadhantriyant/pockmon/internal/database"
	"github.com/ramadhantriyant/pockmon/internal/middleware"
	"github.com/ramadhantriyant/pockmon/internal/util"
)

func (h *Handler) ListBudgets(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgets, err := h.config.Querier.ListBudgetsByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"budgets": budgets,
	})
}

func (h *Handler) ListActiveBudgets(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgets, err := h.config.Querier.ListActiveBudgetsByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"budgets": budgets,
	})
}

func (h *Handler) ListBudgetsWithSpending(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	startDate, err := util.GetDate(c.Query("start_date"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid or missing start_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	endDate, err := util.GetDate(c.Query("end_date"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid or missing end_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgets, err := h.config.Querier.ListBudgetsWithSpending(c.Request.Context(), database.ListBudgetsWithSpendingParams{
		UserID:            user.ID,
		TransactionDate:   startDate,
		TransactionDate_2: endDate,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"budgets": budgets,
	})
}

func (h *Handler) ListBudgetsExceedingThreshold(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	startDate, err := util.GetDate(c.Query("start_date"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid or missing start_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	endDate, err := util.GetDate(c.Query("end_date"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid or missing end_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgets, err := h.config.Querier.ListBudgetsExceedingThreshold(c.Request.Context(), database.ListBudgetsExceedingThresholdParams{
		UserID:            user.ID,
		TransactionDate:   startDate,
		TransactionDate_2: endDate,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"budgets": budgets,
	})
}

func (h *Handler) GetBudget(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgetID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid budget id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budget, err := h.config.Querier.GetBudgetByID(c.Request.Context(), database.GetBudgetByIDParams{
		ID:     budgetID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "budget not found"),
				Type: gin.ErrorTypePublic,
			})
		} else {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"budget": budget,
	})
}

func (h *Handler) GetBudgetWithSpending(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgetID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid budget id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	startDate, err := util.GetDate(c.Query("start_date"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid or missing start_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	endDate, err := util.GetDate(c.Query("end_date"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid or missing end_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budget, err := h.config.Querier.GetBudgetWithSpending(c.Request.Context(), database.GetBudgetWithSpendingParams{
		ID:                budgetID,
		UserID:            user.ID,
		TransactionDate:   startDate,
		TransactionDate_2: endDate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "budget not found"),
				Type: gin.ErrorTypePublic,
			})
		} else {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"budget": budget,
	})
}

func (h *Handler) CreateBudget(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var req struct {
		CategoryID     string   `json:"category_id" binding:"required"`
		Name           string   `json:"name" binding:"required"`
		Amount         float64  `json:"amount" binding:"required,gt=0"`
		Period         string   `json:"period" binding:"required,oneof=daily weekly monthly quarterly yearly"`
		StartDate      string   `json:"start_date" binding:"required"`
		EndDate        *string  `json:"end_date"`
		AlertThreshold *float64 `json:"alert_threshold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	categoryID, err := util.GetUUID(req.CategoryID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid category id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err := h.config.Querier.GetCategoryByIDAndUser(c.Request.Context(), database.GetCategoryByIDAndUserParams{
		ID:     categoryID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "category not found"),
				Type: gin.ErrorTypePublic,
			})
		} else {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
		}
		c.Abort()
		return
	}

	startDate, err := util.GetDate(req.StartDate)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid start_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var endDate pgtype.Date
	if req.EndDate != nil {
		endDate, err = util.GetDate(*req.EndDate)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid end_date (use YYYY-MM-DD)"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
	}

	alertThresholdVal := 80.0
	if req.AlertThreshold != nil {
		alertThresholdVal = *req.AlertThreshold
	}
	alertThreshold, err := util.GetNumeric(alertThresholdVal)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	amount, err := util.GetNumeric(req.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgetID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budget, err := h.config.Querier.CreateBudget(c.Request.Context(), database.CreateBudgetParams{
		ID:             budgetID,
		UserID:         user.ID,
		CategoryID:     categoryID,
		Name:           req.Name,
		Amount:         amount,
		Period:         req.Period,
		StartDate:      startDate,
		EndDate:        endDate,
		AlertThreshold: alertThreshold,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated,
		"budget": budget,
	})
}

func (h *Handler) UpdateBudget(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgetID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid budget id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var req struct {
		Name           string   `json:"name" binding:"required"`
		Amount         float64  `json:"amount" binding:"required,gt=0"`
		Period         string   `json:"period" binding:"required,oneof=daily weekly monthly quarterly yearly"`
		StartDate      string   `json:"start_date" binding:"required"`
		EndDate        *string  `json:"end_date"`
		AlertThreshold *float64 `json:"alert_threshold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	startDate, err := util.GetDate(req.StartDate)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid start_date (use YYYY-MM-DD)"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var endDate pgtype.Date
	if req.EndDate != nil {
		endDate, err = util.GetDate(*req.EndDate)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid end_date (use YYYY-MM-DD)"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
	}

	alertThresholdVal := 80.0
	if req.AlertThreshold != nil {
		alertThresholdVal = *req.AlertThreshold
	}
	alertThreshold, err := util.GetNumeric(alertThresholdVal)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	amount, err := util.GetNumeric(req.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budget, err := h.config.Querier.UpdateBudget(c.Request.Context(), database.UpdateBudgetParams{
		ID:             budgetID,
		UserID:         user.ID,
		Name:           req.Name,
		Amount:         amount,
		Period:         req.Period,
		StartDate:      startDate,
		EndDate:        endDate,
		AlertThreshold: alertThreshold,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "budget not found"),
				Type: gin.ErrorTypePublic,
			})
		} else {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"budget": budget,
	})
}

func (h *Handler) DeactivateBudget(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	budgetID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid budget id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err := h.config.Querier.GetBudgetByID(c.Request.Context(), database.GetBudgetByIDParams{
		ID:     budgetID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "budget not found"),
				Type: gin.ErrorTypePublic,
			})
		} else {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
		}
		c.Abort()
		return
	}

	if err := h.config.Querier.DeactivateBudget(c.Request.Context(), database.DeactivateBudgetParams{
		ID:     budgetID,
		UserID: user.ID,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}
