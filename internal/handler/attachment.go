package handler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/ramadhantriyant/pockmon/internal/database"
	"github.com/ramadhantriyant/pockmon/internal/middleware"
	"github.com/ramadhantriyant/pockmon/internal/util"
)

func (h *Handler) signedURL(objectPath, method string, expiry time.Duration) (string, error) {
	return h.config.StorageClient.Bucket(h.config.StorageBucket).SignedURL(objectPath, &storage.SignedURLOptions{
		Method:  method,
		Expires: time.Now().Add(expiry),
		Scheme:  storage.SigningSchemeV4,
	})
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func (h *Handler) GetUploadURL(c *gin.Context) {
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

	if _, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     transactionID,
		UserID: user.ID,
	}); err != nil {
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

	fileName := c.Query("file_name")
	if fileName == "" {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "file_name query parameter is required"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	fileID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	fileIDStr := fmt.Sprintf("%x", fileID.Bytes)
	objectPath := fmt.Sprintf("attachments/%x/%s-%s", transactionID.Bytes, fileIDStr[:8], sanitizeFileName(fileName))

	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(15 * time.Minute),
		Scheme:  storage.SigningSchemeV4,
	}
	if fileType := c.Query("file_type"); fileType != "" {
		opts.ContentType = fileType
	}

	uploadURL, err := h.config.StorageClient.Bucket(h.config.StorageBucket).SignedURL(objectPath, opts)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "failed to generate upload URL").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      http.StatusOK,
		"upload_url":  uploadURL,
		"object_path": objectPath,
		"expires_at":  time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339),
	})
}

func (h *Handler) ConfirmAttachment(c *gin.Context) {
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

	if _, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     transactionID,
		UserID: user.ID,
	}); err != nil {
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

	var req struct {
		FileName string  `json:"file_name" binding:"required"`
		FilePath string  `json:"file_path" binding:"required"`
		FileType *string `json:"file_type"`
		FileSize *int32  `json:"file_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid request body"),
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

	attachment, err := h.config.Querier.CreateAttachment(c.Request.Context(), database.CreateAttachmentParams{
		ID:            id,
		TransactionID: transactionID,
		FileName:      req.FileName,
		FilePath:      req.FilePath,
		FileType:      req.FileType,
		FileSize:      req.FileSize,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	downloadURL, err := h.signedURL(attachment.FilePath, "GET", time.Hour)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "failed to generate download URL").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":       http.StatusCreated,
		"attachment":   attachment,
		"download_url": downloadURL,
	})
}

func (h *Handler) ListAttachmentsByTransaction(c *gin.Context) {
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

	if _, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     transactionID,
		UserID: user.ID,
	}); err != nil {
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

	attachments, err := h.config.Querier.ListAttachmentsByTransaction(c.Request.Context(), transactionID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	type attachmentWithURL struct {
		database.Attachment
		DownloadURL string `json:"download_url"`
	}

	result := make([]attachmentWithURL, 0, len(attachments))
	for _, a := range attachments {
		url, err := h.signedURL(a.FilePath, "GET", time.Hour)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "failed to generate download URL").WithInternal(err.Error()),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		result = append(result, attachmentWithURL{Attachment: a, DownloadURL: url})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      http.StatusOK,
		"attachments": result,
	})
}

func (h *Handler) GetAttachment(c *gin.Context) {
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

	attachmentID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid attachment id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	attachment, err := h.config.Querier.GetAttachmentByID(c.Request.Context(), attachmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "attachment not found"),
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

	// Verify ownership through the transaction.
	if _, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     attachment.TransactionID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "attachment not found"),
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

	downloadURL, err := h.signedURL(attachment.FilePath, "GET", time.Hour)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "failed to generate download URL").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       http.StatusOK,
		"attachment":   attachment,
		"download_url": downloadURL,
	})
}

func (h *Handler) DeleteAttachment(c *gin.Context) {
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

	attachmentID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid attachment id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	attachment, err := h.config.Querier.GetAttachmentByID(c.Request.Context(), attachmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "attachment not found"),
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

	// Verify ownership through the transaction.
	if _, err := h.config.Querier.GetTransactionByID(c.Request.Context(), database.GetTransactionByIDParams{
		ID:     attachment.TransactionID,
		UserID: user.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "attachment not found"),
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

	if err := h.config.StorageClient.Bucket(h.config.StorageBucket).Object(attachment.FilePath).Delete(c.Request.Context()); err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "failed to delete file").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := h.config.Querier.DeleteAttachment(c.Request.Context(), attachmentID); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error").WithInternal(err.Error()),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent)
}
