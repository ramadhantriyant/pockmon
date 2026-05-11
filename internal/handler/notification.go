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

func (h *Handler) ListNotifications(c *gin.Context) {
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

	notifications, err := h.config.Querier.ListNotificationsByUser(c.Request.Context(), database.ListNotificationsByUserParams{
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

	unreadCount, err := h.config.Querier.CountUnreadNotifications(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        http.StatusOK,
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

func (h *Handler) ListUnreadNotifications(c *gin.Context) {
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

	notifications, err := h.config.Querier.ListUnreadNotificationsByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        http.StatusOK,
		"notifications": notifications,
		"unread_count":  len(notifications),
	})
}

func (h *Handler) MarkNotificationAsRead(c *gin.Context) {
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

	notificationID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid notification id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err := h.config.Querier.GetNotificationByID(c.Request.Context(), database.GetNotificationByIDParams{
		ID:     notificationID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "notification not found"),
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

	if err := h.config.Querier.MarkNotificationAsRead(c.Request.Context(), database.MarkNotificationAsReadParams{
		ID:     notificationID,
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

func (h *Handler) MarkAllNotificationsAsRead(c *gin.Context) {
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

	if err := h.config.Querier.MarkAllNotificationsAsRead(c.Request.Context(), user.ID); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteNotification(c *gin.Context) {
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

	notificationID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid notification id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if _, err := h.config.Querier.GetNotificationByID(c.Request.Context(), database.GetNotificationByIDParams{
		ID:     notificationID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "notification not found"),
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

	if err := h.config.Querier.DeleteNotification(c.Request.Context(), database.DeleteNotificationParams{
		ID:     notificationID,
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

func (h *Handler) DeleteReadNotifications(c *gin.Context) {
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

	if err := h.config.Querier.DeleteReadNotifications(c.Request.Context(), user.ID); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}
