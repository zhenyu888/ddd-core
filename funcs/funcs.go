package funcs

import (
	"encoding/base64"
	"reflect"
	"runtime"
	"strings"
	"time"
)

func Base64(str string) string {
	input := []byte(str)
	return base64.StdEncoding.EncodeToString(input)
}

func CurFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	nameSlice := strings.Split(f.Name(), ".")
	return nameSlice[len(nameSlice)-1]
}

func ReflectValue(val interface{}) reflect.Value {
	var rlt reflect.Value
	if reflect.TypeOf(val).Kind() == reflect.Ptr {
		rlt = reflect.ValueOf(val).Elem()
	} else {
		rlt = reflect.ValueOf(val)
	}
	return rlt
}

func TypeEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return false
	}

	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	return v1.Type() == v2.Type()
}

func ReflectValueName(val interface{}) string {
	v := ReflectValue(val)
	name := v.String()
	if !strings.HasPrefix(name, "<") {
		return name
	}
	if strings.Contains(name, ".") {
		pointIdx := strings.LastIndex(name, ".")
		return name[pointIdx+1 : len(name)-7]
	}
	return name[1 : len(name)-7]
}

// DeepCopy copied from https://github.com/mohae/deepcopy/blob/master/deepcopy.go
func DeepCopy(src interface{}) interface{} {
	if src == nil {
		return nil
	}
	ori := ReflectValue(src)
	dst := reflect.New(ori.Type()).Elem()
	copyRecursive(ori, dst)
	return dst.Addr().Interface()
}

func copyRecursive(original, cpy reflect.Value) {
	// handle according to original's Kind
	switch original.Kind() {
	case reflect.Ptr:
		// Get the actual value being pointed to.
		originalValue := original.Elem()

		// if  it isn't valid, return.
		if !originalValue.IsValid() {
			return
		}
		cpy.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, cpy.Elem())

	case reflect.Interface:
		// If this is a nil, don't do anything
		if original.IsNil() {
			return
		}
		// Get the value for the interface, not the pointer.
		originalValue := original.Elem()

		// Get the value by calling Elem().
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		cpy.Set(copyValue)

	case reflect.Struct:
		t, ok := original.Interface().(time.Time)
		if ok {
			cpy.Set(reflect.ValueOf(t))
			return
		}
		// Go through each field of the struct and copy it.
		for i := 0; i < original.NumField(); i++ {
			// The Type's StructField for a given field is checked to see if StructField.PkgPath
			// is set to determine if the field is exported or not because CanSet() returns false
			// for settable fields.  I'm not sure why.  -mohae
			if original.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(original.Field(i), cpy.Field(i))
		}

	case reflect.Slice:
		if original.IsNil() {
			return
		}
		// Make a new slice and copy each element.
		cpy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			copyRecursive(original.Index(i), cpy.Index(i))
		}

	case reflect.Map:
		if original.IsNil() {
			return
		}
		cpy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			copyKey := DeepCopy(key.Interface())
			cpy.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}

	default:
		cpy.Set(original)
	}
}
