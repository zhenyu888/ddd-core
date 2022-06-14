package apperr

import (
	"github.com/pkg/errors"
)

func HandleAppErr(err error, callback ...func(err AppError)) {
	switch err := errors.Cause(err).(type) {
	case AppError:
		if len(callback) > 0 {
			for _, cb := range callback {
				cb(err)
			}
		}
	}
}
