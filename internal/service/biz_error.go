package service

import (
	"errors"
	"fmt"
)

var (
	ErrBadRequest     = errors.New("错误请求")
	ErrNotFound       = errors.New("未找到")
	ErrConflict       = errors.New("冲突")
	ErrInternal       = errors.New("内部错误")
	ErrUnauthorized   = errors.New("未授权")
	ErrForbidden      = errors.New("禁止访问")
	ErrNotImplemented = errors.New("未实现")
)

func BadRequest(msg string) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, msg)
}

func NotFound(msg string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, msg)
}

func Conflict(msg string) error {
	return fmt.Errorf("%w: %s", ErrConflict, msg)
}

func Internal(msg string) error {
	return fmt.Errorf("%w: %s", ErrInternal, msg)
}

func Unauthorized(msg string) error {
	return fmt.Errorf("%w: %s", ErrUnauthorized, msg)
}

func Forbidden(msg string) error {
	return fmt.Errorf("%w: %s", ErrForbidden, msg)
}

func NotImplemented(msg string) error {
	return fmt.Errorf("%w: %s", ErrNotImplemented, msg)
}
