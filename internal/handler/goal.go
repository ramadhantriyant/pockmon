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

func (h *Handler) ListGoals(c *gin.Context) {
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

	goals, err := h.config.Querier.ListGoalsByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"goals":  goals,
	})
}

func (h *Handler) ListActiveGoals(c *gin.Context) {
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

	goals, err := h.config.Querier.ListActiveGoalsByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"goals":  goals,
	})
}

func (h *Handler) ListGoalsByType(c *gin.Context) {
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

	goalType := c.Param("type")
	validTypes := map[string]bool{
		"savings": true, "debt_payoff": true, "investment": true,
		"purchase": true, "other": true,
	}
	if !validTypes[goalType] {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid goal type, must be one of: savings, debt_payoff, investment, purchase, other"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	goals, err := h.config.Querier.ListGoalsByType(c.Request.Context(), database.ListGoalsByTypeParams{
		UserID:   user.ID,
		GoalType: goalType,
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
		"status": http.StatusOK,
		"goals":  goals,
	})
}

func (h *Handler) GetGoal(c *gin.Context) {
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

	goalID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid goal id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	goal, err := h.config.Querier.GetGoalByID(c.Request.Context(), database.GetGoalByIDParams{
		ID:     goalID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "goal not found"),
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
		"goal":   goal,
	})
}

func (h *Handler) GetGoalProgress(c *gin.Context) {
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

	goalID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid goal id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	progress, err := h.config.Querier.GetGoalProgress(c.Request.Context(), database.GetGoalProgressParams{
		ID:     goalID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "goal not found"),
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
		"status":   http.StatusOK,
		"progress": progress,
	})
}

func (h *Handler) CreateGoal(c *gin.Context) {
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
		AccountID    *string  `json:"account_id"`
		Name         string   `json:"name" binding:"required"`
		Description  *string  `json:"description"`
		TargetAmount float64  `json:"target_amount" binding:"required,gt=0"`
		CurrencyCode *string  `json:"currency_code"`
		TargetDate   *string  `json:"target_date"`
		GoalType     string   `json:"goal_type" binding:"required,oneof=savings debt_payoff investment purchase other"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var accountID pgtype.UUID
	if req.AccountID != nil {
		accountID, err = util.GetUUID(*req.AccountID)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid account id"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		if _, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
			ID:     accountID,
			UserID: user.ID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.Error(&gin.Error{
					Err:  middleware.NewAppError(http.StatusNotFound, "not found", "account not found"),
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
	}

	var targetDate pgtype.Date
	if req.TargetDate != nil {
		targetDate, err = util.GetDate(*req.TargetDate)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid target_date (use YYYY-MM-DD)"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
	}

	targetAmount, err := util.GetNumeric(req.TargetAmount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	id, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	goal, err := h.config.Querier.CreateGoal(c.Request.Context(), database.CreateGoalParams{
		ID:           id,
		UserID:       user.ID,
		AccountID:    accountID,
		Name:         req.Name,
		Description:  req.Description,
		TargetAmount: targetAmount,
		CurrencyCode: req.CurrencyCode,
		TargetDate:   targetDate,
		GoalType:     req.GoalType,
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
		"goal":   goal,
	})
}

func (h *Handler) UpdateGoal(c *gin.Context) {
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

	goalID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid goal id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var req struct {
		AccountID    *string  `json:"account_id"`
		Name         string   `json:"name" binding:"required"`
		Description  *string  `json:"description"`
		TargetAmount float64  `json:"target_amount" binding:"required,gt=0"`
		TargetDate   *string  `json:"target_date"`
		GoalType     string   `json:"goal_type" binding:"required,oneof=savings debt_payoff investment purchase other"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var accountID pgtype.UUID
	if req.AccountID != nil {
		accountID, err = util.GetUUID(*req.AccountID)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid account id"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		if _, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
			ID:     accountID,
			UserID: user.ID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.Error(&gin.Error{
					Err:  middleware.NewAppError(http.StatusNotFound, "not found", "account not found"),
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
	}

	var targetDate pgtype.Date
	if req.TargetDate != nil {
		targetDate, err = util.GetDate(*req.TargetDate)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid target_date (use YYYY-MM-DD)"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
	}

	targetAmount, err := util.GetNumeric(req.TargetAmount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	goal, err := h.config.Querier.UpdateGoal(c.Request.Context(), database.UpdateGoalParams{
		ID:           goalID,
		UserID:       user.ID,
		AccountID:    accountID,
		Name:         req.Name,
		Description:  req.Description,
		TargetAmount: targetAmount,
		TargetDate:   targetDate,
		GoalType:     req.GoalType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "goal not found"),
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
		"goal":   goal,
	})
}

func (h *Handler) ContributeToGoal(c *gin.Context) {
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

	goalID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid goal id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	goal, err := h.config.Querier.GetGoalByID(c.Request.Context(), database.GetGoalByIDParams{
		ID:     goalID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "goal not found"),
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

	if goal.IsCompleted != nil && *goal.IsCompleted {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "goal is already completed"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	currentAmt, err := goal.CurrentAmount.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	newAmount, err := util.GetNumeric(currentAmt.Float64 + req.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	updated, err := h.config.Querier.UpdateGoalCurrentAmount(c.Request.Context(), database.UpdateGoalCurrentAmountParams{
		ID:            goalID,
		UserID:        user.ID,
		CurrentAmount: newAmount,
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
		"status": http.StatusOK,
		"goal":   updated,
	})
}

func (h *Handler) CompleteGoal(c *gin.Context) {
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

	goalID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid goal id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	goal, err := h.config.Querier.CompleteGoal(c.Request.Context(), database.CompleteGoalParams{
		ID:     goalID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "goal not found"),
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
		"goal":   goal,
	})
}

func (h *Handler) DeleteGoal(c *gin.Context) {
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

	goalID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid goal id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err := h.config.Querier.GetGoalByID(c.Request.Context(), database.GetGoalByIDParams{
		ID:     goalID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "goal not found"),
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

	if err := h.config.Querier.DeleteGoal(c.Request.Context(), database.DeleteGoalParams{
		ID:     goalID,
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
