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

func (h *Handler) ListAccounts(c *gin.Context) {
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

	accounts, err := h.config.Querier.ListActiveAccountsByUser(c.Request.Context(), user.ID)
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
		"account": accounts,
	})
}

func (h *Handler) GetAccount(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"account": account,
	})
}

func (h *Handler) CreateAccount(c *gin.Context) {
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

	var newAccount struct {
		Name           string  `json:"name" binding:"required"`
		Type           string  `json:"type" binding:"required,oneof=cash bank credit_card debit_card investment loan"`
		InitialBalance float64 `json:"initial_balance"`
		IncludeInTotal *bool   `json:"include_in_total"`
		Color          *string `json:"color"`
		Icon           *string `json:"icon"`
		Notes          *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&newAccount); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	accountID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	initialBalance, err := util.GetNumeric(newAccount.InitialBalance)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	params := database.CreateAccountParams{
		ID:             accountID,
		UserID:         user.ID,
		Name:           newAccount.Name,
		Type:           newAccount.Type,
		CurrencyCode:   &user.CurrencyCode,
		InitialBalance: initialBalance,
		IncludeInTotal: newAccount.IncludeInTotal,
		Color:          newAccount.Color,
		Icon:           newAccount.Icon,
		Notes:          newAccount.Notes,
	}

	account, err := h.config.Querier.CreateAccount(c.Request.Context(), params)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"account": account,
	})
}

func (h *Handler) UpdateAccount(c *gin.Context) {
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

	var editedAccount struct {
		Name            string  `json:"name" binding:"required"`
		Type            string  `json:"type" binding:"required,oneof=cash bank credit_card debit_card investment loan"`
		CurrencyCode    *string `json:"currency_code"`
		IncludedInTotal *bool   `json:"include_in_total"`
		Color           *string `json:"color"`
		Icon            *string `json:"icon"`
		Notes           *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&editedAccount); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	// Verify ownership — GetAccountByID filters by both id and user_id
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

	params := database.UpdateAccountParams{
		ID:             accountID,
		Name:           editedAccount.Name,
		Type:           editedAccount.Type,
		CurrencyCode:   editedAccount.CurrencyCode,
		IncludeInTotal: editedAccount.IncludedInTotal,
		Color:          editedAccount.Color,
		Icon:           editedAccount.Icon,
		Notes:          editedAccount.Notes,
		UserID:         user.ID,
	}
	account, err := h.config.Querier.UpdateAccount(c.Request.Context(), params)
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
		"account": account,
	})
}

func (h *Handler) DeactivateAccount(c *gin.Context) {
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

	// Verify ownership — GetAccountByID filters by both id and user_id
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

	params := database.DeactivateAccountParams{
		ID:     accountID,
		UserID: user.ID,
	}
	if err := h.config.Querier.DeactivateAccount(c.Request.Context(), params); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}
