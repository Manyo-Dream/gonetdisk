package service

import (
	"errors"
	"fmt"
)

var (
	ErrBadRequest = errors.New("错误请求")
	ErrNotFound   = errors.New("未找到")
	ErrConflict   = errors.New("冲突")
	ErrInternal   = errors.New("内部错误")
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