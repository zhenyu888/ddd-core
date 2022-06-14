package ddd

import (
	"strings"

	"github.com/zhenyu888/ddd-core/funcs"
)

/*
Component 组件，一个 Service、Repository、Factory 都可以认为是一个组件
*/
type Component interface {
	Name() string
}

func componentName(c interface{}) string {
	if nameStr, ok := c.(string); ok {
		return nameStr
	}
	if cc, ok := c.(Component); ok {
		return cc.Name()
	}
	v := funcs.ReflectValue(c)
	name := v.String()
	if !strings.HasPrefix(name, "<") {
		return name
	}
	return name[1 : len(name)-7]
}
