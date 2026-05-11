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
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "error generating uuid").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var req struct {
		CurrencyCode string `json:"currency_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "currency_code is required"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	user, err := h.config.Querier.CreateUser(c.Request.Context(), database.CreateUserParams{
		ID:           userID,
		FirebaseUid:  token.UID,
		CurrencyCode: req.CurrencyCode,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user already exists or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	countUser, err := h.config.Querier.CountUser(c.Request.Context())
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	isAdmin := user.IsAdmin
	if countUser == 1 {
		if err := h.config.Querier.SetUserAdmin(c.Request.Context(), database.SetUserAdminParams{
			ID:      userID,
			IsAdmin: true,
		}); err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		isAdmin = true
	}

	if err := util.SeedCategory(c.Request.Context(), h.config.Querier, user.ID); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
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
			"is_admin":      isAdmin,
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
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
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
			"is_admin":      user.IsAdmin,
		},
	})
}

func (h *Handler) UpdateCurrency(c *gin.Context) {
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
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	code := c.Param("code")
	if len(code) < 2 || len(code) > 4 {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid currency code"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err = h.config.Querier.UpdateUserCurrency(c.Request.Context(), database.UpdateUserCurrencyParams{
		ID:           user.ID,
		CurrencyCode: code,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
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
			"currency_code": code,
			"is_admin":      user.IsAdmin,
		},
	})
}
