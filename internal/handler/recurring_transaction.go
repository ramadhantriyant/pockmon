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

func (h *Handler) ListRecurringTransactions(c *gin.Context) {
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

	transactions, err := h.config.Querier.ListRecurringTransactionsByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       http.StatusOK,
		"transactions": transactions,
	})
}

func (h *Handler) ListActiveRecurringTransactions(c *gin.Context) {
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

	transactions, err := h.config.Querier.ListActiveRecurringTransactionsByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       http.StatusOK,
		"transactions": transactions,
	})
}

func (h *Handler) GetRecurringTransaction(c *gin.Context) {
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

	recurringID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid recurring transaction id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transaction, err := h.config.Querier.GetRecurringTransactionByID(c.Request.Context(), database.GetRecurringTransactionByIDParams{
		ID:     recurringID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "recurring transaction not found"),
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
		"status":      http.StatusOK,
		"transaction": transaction,
	})
}

func (h *Handler) CreateRecurringTransaction(c *gin.Context) {
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
		AccountID    string   `json:"account_id" binding:"required"`
		CategoryID   *string  `json:"category_id"`
		Type         string   `json:"type" binding:"required,oneof=expense income"`
		Amount       float64  `json:"amount" binding:"required,gt=0"`
		CurrencyCode *string  `json:"currency_code"`
		Description  string   `json:"description" binding:"required"`
		Frequency    string   `json:"frequency" binding:"required,oneof=daily weekly biweekly monthly quarterly yearly"`
		StartDate    string   `json:"start_date" binding:"required"`
		EndDate      *string  `json:"end_date"`
		AutoCreate   *bool    `json:"auto_create"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	accountID, err := util.GetUUID(req.AccountID)
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

	var categoryID pgtype.UUID
	if req.CategoryID != nil {
		categoryID, err = util.GetUUID(*req.CategoryID)
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

	amount, err := util.GetNumeric(req.Amount)
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

	transaction, err := h.config.Querier.CreateRecurringTransaction(c.Request.Context(), database.CreateRecurringTransactionParams{
		ID:           id,
		UserID:       user.ID,
		AccountID:    accountID,
		CategoryID:   categoryID,
		Type:         req.Type,
		Amount:       amount,
		CurrencyCode: req.CurrencyCode,
		Description:  req.Description,
		Frequency:    req.Frequency,
		StartDate:    startDate,
		EndDate:      endDate,
		NextDueDate:  startDate,
		AutoCreate:   req.AutoCreate,
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
		"status":      http.StatusCreated,
		"transaction": transaction,
	})
}

func (h *Handler) UpdateRecurringTransaction(c *gin.Context) {
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

	recurringID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid recurring transaction id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var req struct {
		AccountID   string  `json:"account_id" binding:"required"`
		CategoryID  *string `json:"category_id"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Description string  `json:"description" binding:"required"`
		Frequency   string  `json:"frequency" binding:"required,oneof=daily weekly biweekly monthly quarterly yearly"`
		EndDate     *string `json:"end_date"`
		AutoCreate  *bool   `json:"auto_create"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	accountID, err := util.GetUUID(req.AccountID)
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

	var categoryID pgtype.UUID
	if req.CategoryID != nil {
		categoryID, err = util.GetUUID(*req.CategoryID)
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

	amount, err := util.GetNumeric(req.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transaction, err := h.config.Querier.UpdateRecurringTransaction(c.Request.Context(), database.UpdateRecurringTransactionParams{
		ID:          recurringID,
		UserID:      user.ID,
		AccountID:   accountID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: req.Description,
		Frequency:   req.Frequency,
		EndDate:     endDate,
		AutoCreate:  req.AutoCreate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "recurring transaction not found"),
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
		"status":      http.StatusOK,
		"transaction": transaction,
	})
}

func (h *Handler) DeactivateRecurringTransaction(c *gin.Context) {
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

	recurringID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid recurring transaction id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err := h.config.Querier.GetRecurringTransactionByID(c.Request.Context(), database.GetRecurringTransactionByIDParams{
		ID:     recurringID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "recurring transaction not found"),
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

	if err := h.config.Querier.DeactivateRecurringTransaction(c.Request.Context(), database.DeactivateRecurringTransactionParams{
		ID:     recurringID,
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
