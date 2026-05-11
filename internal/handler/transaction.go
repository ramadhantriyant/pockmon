package handler

import (
	"errors"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/ramadhantriyant/pockmon/internal/database"
	"github.com/ramadhantriyant/pockmon/internal/middleware"
	"github.com/ramadhantriyant/pockmon/internal/util"
)

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

func (h *Handler) GetTransaction(c *gin.Context) {
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

	transactionID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transaction id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transaction, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     transactionID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "transaction not found"),
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

	var req struct {
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
	if err := c.ShouldBindJSON(&req); err != nil {
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

	accountID, err := util.GetUUID(req.AccountID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid account id"),
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

	amount, err := util.GetNumeric(req.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid amount").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transactionDate, err := util.GetDate(req.TransactionDate)
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
		Type:            req.Type,
		Amount:          amount,
		CurrencyCode:    req.CurrencyCode,
		TransactionDate: transactionDate,
		Description:     req.Description,
		Notes:           req.Notes,
		Payee:           req.Payee,
		Location:        req.Location,
		Tags:            req.Tags,
		IsRecurring:     req.IsRecurring,
	}

	if req.RecurringTransactionID != nil {
		recurringID, err := util.GetUUID(*req.RecurringTransactionID)
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
	if req.Type == "income" {
		newBalFloat = currentBal.Float64 + req.Amount
	} else {
		newBalFloat = currentBal.Float64 - req.Amount
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

func (h *Handler) UpdateTransaction(c *gin.Context) {
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

	transactionID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transaction id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var req struct {
		CategoryID      string   `json:"category_id" binding:"required"`
		Amount          float64  `json:"amount" binding:"required"`
		TransactionDate string   `json:"transaction_date" binding:"required"`
		Description     string   `json:"description" binding:"required"`
		Notes           *string  `json:"notes"`
		Payee           *string  `json:"payee"`
		Location        *string  `json:"location"`
		Tags            []string `json:"tags"`
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

	existing, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     transactionID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "transaction not found"),
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

	account, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
		ID:     existing.AccountID,
		UserID: user.ID,
	})
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
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid amount").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transactionDate, err := util.GetDate(req.TransactionDate)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transaction date").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Compute new balance: reverse the old amount then apply the new one.
	// income: balance += amount on create, so new = current - oldAmt + newAmt
	// expense: balance -= amount on create, so new = current + oldAmt - newAmt
	currentBal, err := account.CurrentBalance.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	oldAmt, err := existing.Amount.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var newBalFloat float64
	if existing.Type == "income" {
		newBalFloat = currentBal.Float64 - oldAmt.Float64 + req.Amount
	} else {
		newBalFloat = currentBal.Float64 + oldAmt.Float64 - req.Amount
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

	// START transaction
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

	transaction, err := q.UpdateTransaction(c.Request.Context(), database.UpdateTransactionParams{
		ID:              transactionID,
		CategoryID:      categoryID,
		Amount:          amount,
		TransactionDate: transactionDate,
		Description:     req.Description,
		Notes:           req.Notes,
		Payee:           req.Payee,
		Location:        req.Location,
		Tags:            req.Tags,
		UserID:          user.ID,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err = q.UpdateAccountBalance(c.Request.Context(), database.UpdateAccountBalanceParams{
		ID:             existing.AccountID,
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
	// END transaction

	c.JSON(http.StatusOK, gin.H{
		"status":      http.StatusOK,
		"transaction": transaction,
	})
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
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

	transactionID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transaction id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	existing, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     transactionID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "transaction not found"),
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

	account, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
		ID:     existing.AccountID,
		UserID: user.ID,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Compute new balance: reverse the old amount then apply the new one.
	// income: balance -= amount on create, so new = current - oldAmt
	// expense: balance += amount on create, so new = current + oldAmt
	currentBal, err := account.CurrentBalance.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	oldAmt, err := existing.Amount.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var newBalFloat float64
	if existing.Type == "income" {
		newBalFloat = currentBal.Float64 - oldAmt.Float64
	} else {
		newBalFloat = currentBal.Float64 + oldAmt.Float64
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

	// START transaction
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

	if err := q.DeleteTransaction(c.Request.Context(), database.DeleteTransactionParams{
		ID:     transactionID,
		UserID: user.ID,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err = q.UpdateAccountBalance(c.Request.Context(), database.UpdateAccountBalanceParams{
		ID:             existing.AccountID,
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
	// START transaction

	c.Status(http.StatusNoContent)
}
