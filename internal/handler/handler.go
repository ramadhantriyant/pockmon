package handler

import "github.com/ramadhantriyant/pockmon/internal/config"

type Handler struct {
	config *config.Config
}

func New(config *config.Config) *Handler {
	return &Handler{
		config: config,
	}
}
