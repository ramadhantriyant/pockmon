package handler

import (
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/ramadhantriyant/pockmon/internal/database"
	"github.com/ramadhantriyant/pockmon/internal/middleware"
	"github.com/ramadhantriyant/pockmon/internal/util"
)

func (h *Handler) ListCategories(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	categories, err := h.config.Querier.ListCategoriesByUser(c.Request.Context(), user.ID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"category": categories,
	})
}

func (h *Handler) GetCategory(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	categoryID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusNotFound, "not found", "invalid category id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}
	category, err := h.config.Querier.GetCategoryByIDAndUser(c.Request.Context(), database.GetCategoryByIDAndUserParams{
		ID:     categoryID,
		UserID: user.ID,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusNotFound, "not found", "invalid category id"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"category": category,
	})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var newCategory struct {
		Name             string  `json:"name" binding:"required"`
		Type             string  `json:"type" binding:"required"`
		Color            *string `json:"color"`
		Icon             *string `json:"icon"`
		ParentCategoryID *string `json:"parent_category_id"`
	}
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid json format"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	categoryID, err := util.GenerateUUID()
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	params := database.CreateCategoryParams{
		ID:     categoryID,
		UserID: user.ID,
		Name:   newCategory.Name,
		Type:   newCategory.Type,
		Color:  newCategory.Color,
		Icon:   newCategory.Icon,
	}

	if newCategory.ParentCategoryID != nil {
		parentCatID, err := util.GetUUID(*newCategory.ParentCategoryID)
		if err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid parent category ID"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		if _, err = h.config.Querier.GetCategoryByID(c.Request.Context(), parentCatID); err != nil {
			c.Error(&gin.Error{
				Err:  middleware.NewAppError(http.StatusNotFound, "not found", "parent category not found"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}
		params.ParentCategoryID = parentCatID
	}

	category, err := h.config.Querier.CreateCategory(c.Request.Context(), params)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":   http.StatusCreated,
		"category": category,
	})
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	var editedCategory struct {
		Name             string `json:"name" binding:"required"`
		Color            string `json:"color"`
		Icon             string `json:"icon"`
		ParentCategoryID string `json:"parent_category_id"`
	}
	if err := c.ShouldBindJSON(&editedCategory); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusBadRequest, "bad request", "invalid json format"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	categoryID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusNotFound, "not found", "category id not found"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	parentCategoryID, err := util.GetUUID(editedCategory.ParentCategoryID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusNotFound, "not found", "parent category ID not found or invalid"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	_, err = h.config.Querier.GetCategoryByID(c.Request.Context(), parentCategoryID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusNotFound, "not found", "invalid parent category ID"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	category, err := h.config.Querier.UpdateCategory(c.Request.Context(), database.UpdateCategoryParams{
		ID:               categoryID,
		Name:             editedCategory.Name,
		Color:            &editedCategory.Color,
		Icon:             &editedCategory.Icon,
		ParentCategoryID: parentCategoryID,
		UserID:           user.ID,
	})
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"category": category,
	})
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	token := c.MustGet("firebaseToken").(*auth.Token)

	user, err := h.config.Querier.GetUserByFirebaseUID(c.Request.Context(), token.UID)
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "user not found or internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	categoryID, err := util.GetUUID(c.Param("id"))
	if err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusNotFound, "not found", "category id not found"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	if err := h.config.Querier.DeleteCategory(c.Request.Context(), database.DeleteCategoryParams{
		ID:     categoryID,
		UserID: user.ID,
	}); err != nil {
		c.Error(&gin.Error{
			Err:  middleware.NewAppError(http.StatusInternalServerError, "internal server error", "internal server error"),
			Type: gin.ErrorTypePublic,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
