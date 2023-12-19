package handler

import (
	"flow-indexer/internal/service"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) Handler {
	return Handler{service}
}
