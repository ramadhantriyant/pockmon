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

func (h *Handler) ListTransfers(c *gin.Context) {
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

	transfers, err := h.config.Querier.ListTransfersByUser(c.Request.Context(), database.ListTransfersByUserParams{
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
		"status":   http.StatusOK,
		"transfer": transfers,
	})
}

func (h *Handler) ListTransfersByAccount(c *gin.Context) {
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

	transfers, err := h.config.Querier.ListTransfersByAccount(c.Request.Context(), database.ListTransfersByAccountParams{
		UserID:        user.ID,
		FromAccountID: accountID,
		Limit:         int32(limit),
		Offset:        int32(offset),
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
		"status":   http.StatusOK,
		"transfer": transfers,
	})
}

func (h *Handler) GetTransfer(c *gin.Context) {
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

	transferID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transfer id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transfer, err := h.config.Querier.GetTransferByID(c.Request.Context(), database.GetTransferByIDParams{
		ID:     transferID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "transfer not found"),
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
		"transfer": transfer,
	})
}

func (h *Handler) CreateTransfer(c *gin.Context) {
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
		FromAccountID string   `json:"from_account_id" binding:"required"`
		ToAccountID   string   `json:"to_account_id" binding:"required"`
		Amount        float64  `json:"amount" binding:"required"`
		ToAmount      *float64 `json:"to_amount"`     // defaults to Amount if same currency
		ExchangeRate  *float64 `json:"exchange_rate"` // defaults to 1.0
		TransferDate  string   `json:"transfer_date" binding:"required"`
		Description   *string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	fromAccountID, err := util.GetUUID(req.FromAccountID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid from_account_id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	toAccountID, err := util.GetUUID(req.ToAccountID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid to_account_id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	fromAccount, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
		ID:     fromAccountID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "from account not found"),
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

	toAccount, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
		ID:     toAccountID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "to account not found"),
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

	// Apply defaults
	toAmountFloat := req.Amount
	if req.ToAmount != nil {
		toAmountFloat = *req.ToAmount
	}
	exchangeRateFloat := 1.0
	if req.ExchangeRate != nil {
		exchangeRateFloat = *req.ExchangeRate
	}

	// Find "Transfer Out" (expense) and "Transfer In" (income) categories.
	// Created on first transfer if they don't exist yet.
	var transferOutCatID, transferInCatID pgtype.UUID
	userCategories, err := h.config.Querier.ListCategoriesByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	for _, cat := range userCategories {
		if cat.Name == "Transfer Out" && cat.Type == "expense" {
			transferOutCatID = cat.ID
		}
		if cat.Name == "Transfer In" && cat.Type == "income" {
			transferInCatID = cat.ID
		}
	}

	// Convert amounts and date before starting the DB transaction
	amount, err := util.GetNumeric(req.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid amount").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	toAmount, err := util.GetNumeric(toAmountFloat)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid to_amount").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	exchangeRate, err := util.GetNumeric(exchangeRateFloat)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	transferDate, err := util.GetDate(req.TransferDate)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transfer date").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Compute new account balances
	fromCurrentBal, err := fromAccount.CurrentBalance.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	toCurrentBal, err := toAccount.CurrentBalance.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	fromNewBalance, err := util.GetNumeric(fromCurrentBal.Float64 - req.Amount)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	toNewBalance, err := util.GetNumeric(toCurrentBal.Float64 + toAmountFloat)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Generate IDs up front
	transferID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	fromTxID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	toTxID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	fromCurrencyCode := user.CurrencyCode
	if fromAccount.CurrencyCode != nil {
		fromCurrencyCode = *fromAccount.CurrencyCode
	}
	toCurrencyCode := user.CurrencyCode
	if toAccount.CurrencyCode != nil {
		toCurrencyCode = *toAccount.CurrencyCode
	}

	fromDesc := "Transfer to " + toAccount.Name
	toDesc := "Transfer from " + fromAccount.Name
	if req.Description != nil {
		fromDesc = *req.Description
		toDesc = *req.Description
	}

	// Atomic DB transaction
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

	// Create "Transfer Out" category if this user doesn't have one yet
	if !transferOutCatID.Valid {
		catID, err := util.GenerateUUID()
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		cat, err := q.CreateCategory(c.Request.Context(), database.CreateCategoryParams{
			ID:     catID,
			UserID: user.ID,
			Name:   "Transfer Out",
			Type:   "expense",
		})
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		if err := q.SetSystemCategory(c.Request.Context(), database.SetSystemCategoryParams{
			ID:     cat.ID,
			UserID: user.ID,
		}); err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		transferOutCatID = cat.ID
	}

	// Create "Transfer In" category if this user doesn't have one yet
	if !transferInCatID.Valid {
		catID, err := util.GenerateUUID()
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		cat, err := q.CreateCategory(c.Request.Context(), database.CreateCategoryParams{
			ID:     catID,
			UserID: user.ID,
			Name:   "Transfer In",
			Type:   "income",
		})
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		if err := q.SetSystemCategory(c.Request.Context(), database.SetSystemCategoryParams{
			ID:     cat.ID,
			UserID: user.ID,
		}); err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		transferInCatID = cat.ID
	}

	// Create from_transaction: expense on the source account
	fromTx, err := q.CreateTransaction(c.Request.Context(), database.CreateTransactionParams{
		ID:              fromTxID,
		UserID:          user.ID,
		AccountID:       fromAccountID,
		CategoryID:      transferOutCatID,
		Type:            "expense",
		Amount:          amount,
		CurrencyCode:    fromCurrencyCode,
		TransactionDate: transferDate,
		Description:     fromDesc,
		Tags:            []string{},
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Create to_transaction: income on the destination account
	toTx, err := q.CreateTransaction(c.Request.Context(), database.CreateTransactionParams{
		ID:              toTxID,
		UserID:          user.ID,
		AccountID:       toAccountID,
		CategoryID:      transferInCatID,
		Type:            "income",
		Amount:          toAmount,
		CurrencyCode:    toCurrencyCode,
		TransactionDate: transferDate,
		Description:     toDesc,
		Tags:            []string{},
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transfer, err := q.CreateTransfer(c.Request.Context(), database.CreateTransferParams{
		ID:                transferID,
		UserID:            user.ID,
		FromAccountID:     fromAccountID,
		ToAccountID:       toAccountID,
		FromTransactionID: fromTx.ID,
		ToTransactionID:   toTx.ID,
		Amount:            amount,
		ToAmount:          toAmount,
		ExchangeRate:      exchangeRate,
		TransferDate:      transferDate,
		Description:       req.Description,
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
		ID:             fromAccountID,
		CurrentBalance: fromNewBalance,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err = q.UpdateAccountBalance(c.Request.Context(), database.UpdateAccountBalanceParams{
		ID:             toAccountID,
		CurrentBalance: toNewBalance,
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
		"status":   http.StatusCreated,
		"transfer": transfer,
	})
}

func (h *Handler) DeleteTransfer(c *gin.Context) {
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

	transferID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid transfer id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	transfer, err := h.config.Querier.GetTransferByID(c.Request.Context(), database.GetTransferByIDParams{
		ID:     transferID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "transfer not found"),
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

	// Get both accounts to compute reversed balances
	fromAccount, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
		ID:     transfer.FromAccountID,
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

	toAccount, err := h.config.Querier.GetAccountByID(c.Request.Context(), database.GetAccountByIDParams{
		ID:     transfer.ToAccountID,
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

	// Reverse the original balance changes
	fromCurrentBal, err := fromAccount.CurrentBalance.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	toCurrentBal, err := toAccount.CurrentBalance.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	transferAmt, err := transfer.Amount.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	transferToAmt, err := transfer.ToAmount.Float64Value()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Transfer creation: from_account -= amount, to_account += to_amount
	// Deletion reversal: from_account += amount, to_account -= to_amount
	fromRestoredBalance, err := util.GetNumeric(fromCurrentBal.Float64 + transferAmt.Float64)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	toRestoredBalance, err := util.GetNumeric(toCurrentBal.Float64 - transferToAmt.Float64)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

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

	// Delete transfer record first, then the two linked transactions
	if err := q.DeleteTransfer(c.Request.Context(), database.DeleteTransferParams{
		ID:     transferID,
		UserID: user.ID,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := q.DeleteTransaction(c.Request.Context(), database.DeleteTransactionParams{
		ID:     transfer.FromTransactionID,
		UserID: user.ID,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := q.DeleteTransaction(c.Request.Context(), database.DeleteTransactionParams{
		ID:     transfer.ToTransactionID,
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
		ID:             transfer.FromAccountID,
		CurrentBalance: fromRestoredBalance,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err = q.UpdateAccountBalance(c.Request.Context(), database.UpdateAccountBalanceParams{
		ID:             transfer.ToAccountID,
		CurrentBalance: toRestoredBalance,
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

	c.Status(http.StatusNoContent)
}
