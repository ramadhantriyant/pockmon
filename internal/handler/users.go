package handler

import (
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/ramadhantriyant/pockmon/internal/database"
	"github.com/ramadhantriyant/pockmon/internal/middleware"
	"github.com/ramadhantriyant/pockmon/internal/util"
)

func (h *Handler) Register(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)
	name, _ := token.Claims["name"].(string)
	photoURL, _ := token.Claims["picture"].(string)
	email, ok := token.Claims["email"].(string)
	if !ok {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "failed to get email claims"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	userID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "error generating uuid"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var currencyCode struct {
		CurrencyCode string `json:"currency_code"`
	}
	if err := c.ShouldBindJSON(&currencyCode); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid json format"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	user, err := h.config.Querier.CreateUser(c.Request.Context(), database.CreateUserParams{
		ID:           userID,
		FirebaseUid:  token.UID,
		CurrencyCode: currencyCode.CurrencyCode,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user already exists or internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := util.SeedCategory(c.Request.Context(), h.config.Querier, user.ID); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated,
		"user": gin.H{
			"email":         email,
			"display_name":  name,
			"photo_url":     photoURL,
			"currency_code": user.CurrencyCode,
		},
	})
}

func (h *Handler) GetMe(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)
	name, _ := token.Claims["name"].(string)
	photoURL, _ := token.Claims["picture"].(string)
	email, ok := token.Claims["email"].(string)
	if !ok {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "failed to get email claims"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"user": gin.H{
			"email":         email,
			"display_name":  name,
			"photo_url":     photoURL,
			"currency_code": user.CurrencyCode,
		},
	})
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := h.config.Querier.DeleteUser(c.Request.Context(), user.ID); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
