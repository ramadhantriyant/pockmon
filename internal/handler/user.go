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

func (h *Handler) ListUsers(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if !user.IsAdmin {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusForbidden, "forbidden", "insufficient permission"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	users, err := h.config.Querier.ListUser(c.Request.Context())
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
		"user":   users,
	})
}

func (h *Handler) ToggleAdmin(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if !user.IsAdmin {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusForbidden, "forbidden", "insufficient permission"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	userID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid user id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	existingUser, err := h.config.Querier.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "user not found"),
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

	if user.ID == existingUser.ID {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusForbidden, "forbidden", "cannot demote yourself"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := h.config.Querier.SetUserAdmin(c.Request.Context(), database.SetUserAdminParams{
		ID:      userID,
		IsAdmin: !existingUser.IsAdmin,
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
			"id":            existingUser.ID,
			"firebase_uid":  existingUser.FirebaseUid,
			"currency_code": existingUser.CurrencyCode,
			"is_admin":      !existingUser.IsAdmin,
			"created_at":    existingUser.CreatedAt,
			"updated_at":    existingUser.UpdatedAt,
		},
	})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if !user.IsAdmin {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusForbidden, "forbidden", "insufficient permission"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	userID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid user id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	existingUser, err := h.config.Querier.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "user not found"),
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

	if user.ID == existingUser.ID {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusForbidden, "forbidden", "cannot delete yourself"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := h.config.Querier.DeleteUser(c.Request.Context(), userID); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := h.config.AuthClient.DeleteUser(c.Request.Context(), existingUser.FirebaseUid); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}
