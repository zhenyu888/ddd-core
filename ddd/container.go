package ddd

import (
	"sync"
)

var (
	componentMap map[string]interface{}
	lockMap      map[string]*sync.Mutex
	lockLock     sync.Mutex
)

// LoadOrStoreComponent 用来保证一个组件是单例的
func LoadOrStoreComponent(c interface{}, constructFn func() interface{}) interface{} {
	name := componentName(c)
	if component, ok := componentMap[name]; ok {
		return component
	}

	lock := getLock(name)
	lock.Lock()
	defer lock.Unlock()

	if component, ok := componentMap[name]; ok {
		return component
	}

	if componentMap == nil {
		componentMap = make(map[string]interface{})
	}
	component := constructFn()
	componentMap[name] = component
	return component
}

func LoadComponent(c interface{}) (interface{}, bool) {
	name := componentName(c)
	rlt, ok := componentMap[name]
	return rlt, ok
}

func MustLoadComponent(c interface{}) interface{} {
	name := componentName(c)
	rlt, ok := componentMap[name]
	if !ok {
		panic("component not found")
	}
	return rlt
}

func getLock(name string) *sync.Mutex {
	if lock, ok := lockMap[name]; ok {
		return lock
	}
	lockLock.Lock()
	defer lockLock.Unlock()

	if lock, ok := lockMap[name]; ok {
		return lock
	}
	if lockMap == nil {
		lockMap = make(map[string]*sync.Mutex)
	}
	lockMap[name] = &sync.Mutex{}
	return lockMap[name]
}
