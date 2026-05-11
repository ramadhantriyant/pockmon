package handler

import (
	"errors"
	"net/http"
	"strconv"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/ramadhantriyant/pockmon/internal/database"
	"github.com/ramadhantriyant/pockmon/internal/middleware"
	"github.com/ramadhantriyant/pockmon/internal/util"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

func parsePagination(c *gin.Context) (limit, offset int) {
	limit = defaultLimit
	offset = 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}

func (h *Handler) ListTransactions(c *gin.Context) {
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

	limit, offset := parsePagination(c)

	transactions, err := h.config.Querier.ListTransactionsByUser(c.Request.Context(), database.ListTransactionsByUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
		Offset: int32(offset),
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
		"status":      http.StatusOK,
		"transaction": transactions,
	})
}

func (h *Handler) ListTransactionsByCategory(c *gin.Context) {
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

	categoryID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid category id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	limit, offset := parsePagination(c)

	transactions, err := h.config.Querier.ListTransactionsByCategory(c.Request.Context(), database.ListTransactionsByCategoryParams{
		UserID:     user.ID,
		CategoryID: categoryID,
		Limit:      int32(limit),
		Offset:     int32(offset),
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
		"status":      http.StatusOK,
		"transaction": transactions,
	})
}

func (h *Handler) ListTransactionsByAccount(c *gin.Context) {
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

	accountID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid account id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	limit, offset := parsePagination(c)

	transactions, err := h.config.Querier.ListTransactionsByAccount(c.Request.Context(), database.ListTransactionsByAccountParams{
		UserID:    user.ID,
		AccountID: accountID,
		Limit:     int32(limit),
		Offset:    int32(offset),
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
		"status":      http.StatusOK,
		"transaction": transactions,
	})
}

func (h *Handler) ListTransactionsByType(c *gin.Context) {
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

	typ := c.Param("type")
	if typ != "income" && typ != "expense" {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "type must be income or expense"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	limit, offset := parsePagination(c)

	transactions, err := h.config.Querier.ListTransactionsByType(c.Request.Context(), database.ListTransactionsByTypeParams{
		UserID: user.ID,
		Type:   typ,
		Limit:  int32(limit),
		Offset: int32(offset),
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
		"status":      http.StatusOK,
		"transaction": transactions,
	})
}

func (h *Handler) ListTransactionsByTags(c *gin.Context) {
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

	tags := c.QueryArray("tags")
	if len(tags) == 0 {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "at least one tag is required"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	limit, offset := parsePagination(c)

	transactions, err := h.config.Querier.ListTransactionsByTag(c.Request.Context(), database.ListTransactionsByTagParams{
		UserID: user.ID,
		Tags:   tags,
		Limit:  int32(limit),
		Offset: int32(offset),
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
		"status":      http.StatusOK,
		"transaction": transactions,
	})
}

func (h *Handler) CreateTransaction(c *gin.Context) {
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

	var newTransaction struct {
		AccountID              string   `json:"account_id" binding:"required"`
		CategoryID             string   `json:"category_id" binding:"required"`
		Type                   string   `json:"type" binding:"required,oneof=income expense"`
		Amount                 float64  `json:"amount" binding:"required"`
		CurrencyCode           string   `json:"currency_code" binding:"required"`
		TransactionDate        string   `json:"transaction_date" binding:"required"`
		Description            string   `json:"description" binding:"required"`
		Notes                  *string  `json:"notes"`
		Payee                  *string  `json:"payee"`
		Location               *string  `json:"location"`
		Tags                   []string `json:"tags"`
		IsRecurring            *bool    `json:"is_recurring"`
		RecurringTransactionID *string  `json:"recurring_transaction_id"`
	}
	if err := c.ShouldBindJSON(&newTransaction); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transactionID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	accountID, err := util.GetUUID(newTransaction.AccountID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid account id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	categoryID, err := util.GetUUID(newTransaction.CategoryID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid category id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Validate account ownership and get current balance
	account, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
		ID:     accountID,
		UserID: user.ID,
	})
	if err != nil {
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

	// Validate category ownership
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

	amount, err := util.GetNumeric(newTransaction.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid amount").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transactionDate, err := util.GetDate(newTransaction.TransactionDate)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transaction date").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	params := database.CreateTransactionParams{
		ID:              transactionID,
		UserID:          user.ID,
		AccountID:       accountID,
		CategoryID:      categoryID,
		Type:            newTransaction.Type,
		Amount:          amount,
		CurrencyCode:    newTransaction.CurrencyCode,
		TransactionDate: transactionDate,
		Description:     newTransaction.Description,
		Notes:           newTransaction.Notes,
		Payee:           newTransaction.Payee,
		Location:        newTransaction.Location,
		Tags:            newTransaction.Tags,
		IsRecurring:     newTransaction.IsRecurring,
	}

	if newTransaction.RecurringTransactionID != nil {
		recurringID, err := util.GetUUID(*newTransaction.RecurringTransactionID)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid recurring transaction id"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		params.RecurringTransactionID = recurringID
	}

	// Compute new account balance
	currentBal, err := account.CurrentBalance.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var newBalFloat float64
	if newTransaction.Type == "income" {
		newBalFloat = currentBal.Float64 + newTransaction.Amount
	} else {
		newBalFloat = currentBal.Float64 - newTransaction.Amount
	}

	newBalance, err := util.GetNumeric(newBalFloat)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Create transaction and update balance atomically
	tx, err := h.config.DB.Begin(c.Request.Context())
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	defer tx.Rollback(c.Request.Context())

	q := database.New(tx)

	transaction, err := q.CreateTransaction(c.Request.Context(), params)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err = q.UpdateAccountBalance(c.Request.Context(), database.UpdateAccountBalanceParams{
		ID:             accountID,
		CurrentBalance: newBalance,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
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

func (h *Handler) UpdateTransaction(c *gin.Context) {}

func (h *Handler) DeleteTransaction(c *gin.Context) {}
