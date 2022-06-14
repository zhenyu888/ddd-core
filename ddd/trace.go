package ddd

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/zhenyu888/ddd-core/diff"
	"github.com/zhenyu888/ddd-core/funcs"
)

const traceTag = "trace"

func Trace(aggregate, snapshot Aggregate) diff.AggregateDiff {
	builder := diff.NewAggregateDiffBuilder()
	if snapshot == nil {
		builder.SetSelfChanged(true)
		return builder.Build()
	}
	v1 := funcs.ReflectValue(aggregate)
	v2 := funcs.ReflectValue(snapshot)
	if v1.Type() != v2.Type() {
		builder.SetSelfChanged(true)
		return builder.Build()
	}
	// slice 和 map 类型的字段先暂时不做diff
	sliceField := make(map[string][]int)
	mapField := make(map[string][]int)
	for i, n := 0, v1.NumField(); i < n; i++ {
		fieldType := v1.Type().Field(i)
		if fieldType.Name == "AggregateManager" || fieldType.Name == "MixModel" {
			continue
		}

		fieldTag := fieldType.Tag.Get(traceTag)
		// 所有没有tag的字段看做一个整体
		if !isValidTag(fieldTag) {
			if !builder.IsSelfChanged() {
				selfChanged := !reflect.DeepEqual(v1.Field(i).Interface(), v2.Field(i).Interface())
				builder.SetSelfChanged(selfChanged)
			}
			continue
		}
		// 所有具有相同tag的字段看做一个整体
		tagName, _ := parseTag(fieldTag)
		if changed := builder.GetDiff(tagName); changed.IsChanged() {
			continue
		}

		fieldV1 := v1.Field(i)
		switch fieldV1.Kind() {
		case reflect.Array, reflect.Slice:
			sliceField[tagName] = append(sliceField[tagName], i)
		case reflect.Map:
			mapField[tagName] = append(mapField[tagName], i)
		default:
			changed := !reflect.DeepEqual(fieldV1.Interface(), v2.Field(i).Interface())
			if changed {
				builder.PutDiff(tagName, diff.NewDiff(changed))
			}
		}
	}
	fillListDiff(builder, v1, v2, sliceField)
	fillListDiff(builder, v1, v2, mapField)
	return builder.Build()
}

func fillListDiff(builder diff.AggregateDiffBuilder, v1, v2 reflect.Value, tagToIdxs map[string][]int) {
	for tagName, indexSlice := range tagToIdxs {
		if len(indexSlice) == 0 {
			continue
		}
		if changed := builder.GetDiff(tagName); changed.IsChanged() {
			// 如果当前tag已经标识为有改动，则不需要再做diff
			if changed.IsChanged() {
				continue
			} else {
				// Slice/Map 字段的tag如果跟其他字段的tag重复了，也不会ListDiff
				for _, idx := range indexSlice {
					if !reflect.DeepEqual(v1.Field(idx).Interface(), v2.Field(idx).Interface()) {
						builder.PutDiff(tagName, diff.NewDiff(true))
						break
					}
				}
			}
		} else {
			if len(indexSlice) > 1 {
				for _, idx := range indexSlice {
					if !reflect.DeepEqual(v1.Field(idx).Interface(), v2.Field(idx).Interface()) {
						builder.PutDiff(tagName, diff.NewDiff(true))
						break
					}
				}
			} else {
				idx := indexSlice[0]
				builder.PutListDiff(tagName, makeListDiff(v1.Field(idx), v2.Field(idx)))
			}
		}
	}
}

func makeListDiff(x, y reflect.Value) diff.ListDiff {
	builder := diff.NewListDiffBuilder()
	switch x.Kind() {
	case reflect.Slice, reflect.Array:
		xMap := make(map[int64]interface{})
		xSlice := make([]interface{}, 0, x.Len())
		yMap := make(map[int64]interface{})
		ySlice := make([]interface{}, 0, y.Len())
		for i := 0; i < x.Len(); i++ {
			idxV := x.Index(i).Interface()
			if e, ok := idxV.(Entity); ok {
				xMap[e.Identifier()] = e
			}
			xSlice = append(xSlice, idxV)
		}
		for i := 0; i < y.Len(); i++ {
			idxV := y.Index(i).Interface()
			if e, ok := idxV.(Entity); ok {
				yMap[e.Identifier()] = e
			}
			ySlice = append(ySlice, idxV)
		}
		if len(xMap) == 0 || len(yMap) == 0 {
			for _, v := range xSlice {
				builder.AppendAdded(v)
			}
			for _, v := range ySlice {
				builder.AppendRemoved(v)
			}
		} else {
			for kX, vX := range xMap {
				if vY, ok := yMap[kX]; !ok {
					builder.AppendAdded(vX)
				} else if !reflect.DeepEqual(vX, vY) {
					builder.AppendModified(vX)
				}
			}
			for kY, vY := range yMap {
				if _, ok := xMap[kY]; !ok {
					builder.AppendRemoved(vY)
				}
			}
		}
	case reflect.Map:
		iterX := x.MapRange()
		for iterX.Next() {
			kX := iterX.Key()
			vX := iterX.Value()
			if vY := y.MapIndex(kX); vY.IsZero() {
				builder.AppendAdded(vX.Interface())
			} else if !reflect.DeepEqual(vX.Interface(), vY.Interface()) {
				builder.AppendModified(vX.Interface())
			}
		}
		iterY := y.MapRange()
		for iterY.Next() {
			kY := iterY.Key()
			vY := iterY.Value()
			if vX := x.MapIndex(kY); vX.IsZero() {
				builder.AppendRemoved(vY.Interface())
			}
		}
	}
	return builder.Build()
}

// copied from json
// tagOptions is the string following a comma in a struct field's "trace"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// copied from json
// parseTag splits a struct field's trace tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, ""
}

// copied from json
// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			return false
		}
	}
	return true
}
