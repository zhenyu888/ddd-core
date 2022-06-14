package apperr

import (
	"fmt"
)

type ErrType string

const (
	ErrTypeOther ErrType = "UNKNOWN_ERROR"
	ErrTypeBiz   ErrType = "BIZ_ERROR"
	ErrTypeSys   ErrType = "SYS_ERROR"
)

type AppError interface {
	error
	Code() int
	Message() string
	Parent() error
	// With 用来保存错误发生时的上下文信息
	// 比如参数错误，可以通过With记录下当时请求的具体参数值 With("params", req)
	With(k string, v interface{}) AppError
	ErrType() ErrType
}

type baseError struct {
	errCode    int
	errMessage string
	parent     error
	errCtx     map[string]interface{}
}

func newBaseError(code int, message string, es ...error) baseError {
	e := baseError{
		errCode:    code,
		errMessage: message,
	}
	if len(es) > 0 {
		e.parent = es[0]
	}
	return e
}

func (b *baseError) With(k string, v interface{}) AppError {
	if b.errCtx == nil {
		b.errCtx = make(map[string]interface{})
	}
	b.errCtx[k] = v
	return b
}

func (b *baseError) Error() string {
	return b.formatError(b.ErrType())
}

func (b *baseError) formatError(errType ErrType) string {
	var content string
	if b.parent == nil {
		content = fmt.Sprintf("[%s-%d] %s", errType, b.errCode, b.errMessage)
	} else {
		content = fmt.Sprintf("[%s-%d] %s, parent error is %v", errType, b.errCode, b.errMessage, b.parent)
	}
	if len(b.errCtx) > 0 {
		content = fmt.Sprintf("%s; ctx is %v", content, b.errCtx)
	}
	return content
}

func (b *baseError) ErrType() ErrType {
	return ErrTypeOther
}

func (b *baseError) Code() int {
	return b.errCode
}

func (b *baseError) Message() string {
	if b.errMessage != "" {
		return b.errMessage
	}
	if b.parent != nil {
		if pe, ok := b.parent.(AppError); ok {
			return pe.Message()
		} else {
			return b.parent.Error()
		}
	}
	return ""
}

func (b *baseError) Parent() error {
	return b.parent
}

type bizError struct {
	baseError
}

func (be *bizError) ErrType() ErrType {
	return ErrTypeBiz
}

func (be *bizError) Error() string {
	return be.formatError(be.ErrType())
}

func (be *bizError) With(k string, v interface{}) AppError {
	_ = be.baseError.With(k, v)
	return be
}

type sysError struct {
	baseError
}

func (se *sysError) ErrType() ErrType {
	return ErrTypeSys
}

func (se *sysError) Error() string {
	return se.formatError(se.ErrType())
}

func (se *sysError) With(k string, v interface{}) AppError {
	_ = se.baseError.With(k, v)
	return se
}
