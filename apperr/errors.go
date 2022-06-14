package apperr

func BizError(code int, message string, parent ...error) AppError {
	return &bizError{baseError: newBaseError(code, message, parent...)}
}

func SysError(code int, message string, parent ...error) AppError {
	return &sysError{baseError: newBaseError(code, message, parent...)}
}

func ErrBadParam(msg string, v interface{}) AppError {
	return BizError(400, msg).With("params", v)
}

func ErrNotFound(msg string, k string, v interface{}) AppError {
	return BizError(404, msg).With(k, v)
}

func ErrDBFail(ori error, msg string) AppError {
	return SysError(500, msg, ori)
}

func ErrRpcFail(ori error, msg string, req interface{}) AppError {
	return SysError(500, msg, ori).With("RpcRequest", req)
}
