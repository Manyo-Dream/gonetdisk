package controller

import (
	"errors"
	"net/http"

	"github.com/manyodream/gonetdisk/internal/service"
)

func statusFromErr(err error) int {
	switch {
	case errors.Is(err, service.ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, service.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, service.ErrForbidden):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
