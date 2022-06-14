package ddd

import (
	"context"
)

// IdGenerator 生成int64类型唯一标识，默认实现在edu/common_infra库里
type IdGenerator interface {
	Gen(ctx context.Context) (int64, error)
}
