package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ramadhantriyant/pockmon/internal/config"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Handler struct {
	config *config.Config
}

func New(config *config.Config) *Handler {
	return &Handler{
		config: config,
	}
}

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
