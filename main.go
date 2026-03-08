package main

import (
	"encoding/base64"
	"fmt"
	OSLrand "math/rand"
	"strconv"
	"strings"
	"bytes"
	"encoding/json"
	"bufio"
	"os"
	"reflect"
	OSLio "io"
	"time"
	"math"
	"runtime"
	"sort"
	"unsafe"
	"sync"
	"path/filepath"
	"net/http"
	OSL_bytes "bytes"
	OSL_image "image"
	"image/png"
	"image/jpeg"
	OSL_draw "golang.org/x/image/draw"
	OSL_img_resize "github.com/nfnt/resize"
	OSL_exif "github.com/rwcarlsen/goexif/exif"
	OSL_color "image/color"
	"github.com/gin-gonic/gin"
	"github.com/rwcarlsen/goexif/exif"
	"crypto/md5"
	"encoding/hex"
)

var OSLwincreatetime float64 = OSLcastNumber(time.Now().UnixMilli())
var OSLsystem_os = runtime.GOOS
var OSLtimer func() float64 = func() float64 { return float64(time.Now().Unix()) - (OSLwincreatetime / 1000) }
var OSLtimestamp func() int64 = func() int64 { return time.Now().UnixMilli() }
var timer float64
var timestamp int64
func OSLupdateTimer() {
	timer = OSLtimer()
	timestamp = OSLtimestamp()
}

// This is a set of funtions that are used in the compiler for OSL.go

func getGamepads() []any {
	// Stub implementation - returns empty array
	// This should be implemented with actual gamepad support
	return []any{}
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

func OSLlen(s any) int {
	if s == nil {
		return 0
	}
	switch s := s.(type) {
	case string:
		return len(s)
	case []any:
		return len(s)
	case []string:
		return len(s)
	case []int:
		return len(s)
	case []float64:
		return len(s)
	case []bool:
		return len(s)
	case []byte:
		return len(s)
	case []OSLio.Reader:
		return len(s)
	}
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return v.Len()
	}
	if v.Kind() == reflect.Map {
		return v.Len()
	}
	if v.Kind() == reflect.String {
		return len(v.String())
	}
	panic("OSLlen, invalid type: " + v.Kind().String())
}

func OSLtoString(s any) string {
	switch s := s.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	case []any:
		return JsonStringify(s)
	case map[string]any, map[string]string, map[string]int, map[string]float64, map[string]bool:
		return JsonStringify(s)
	case OSLio.Reader:
		data, err := OSLio.ReadAll(s)
		if err != nil {
			panic("OSLcastString: failed to read OSLio.Reader:" + err.Error())
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", s)
	}
}

func OSLcastObject(s any) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	obj, ok := s.(map[string]any)
	if ok {
		return obj
	}
	panic("OSLcastObject, invalid type: " + reflect.TypeOf(s).String())

}

func OSLcastArray(values ...any) []any {
	if len(values) == 1 {
		v := values[0]

		if arr, ok := v.([]any); ok {
			return arr
		}

		rv := reflect.ValueOf(v)

		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return []any{}
			}
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			out := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				out[i] = rv.Index(i).Interface()
			}
			return out
		}

		return []any{v}
	}

	return values
}

func OSLequal(a any, b any) bool {
	if a == b {
		return true
	}
	return strings.EqualFold(OSLtoString(a), OSLtoString(b))
}

func OSLnotEqual(a any, b any) bool {
	if a == b {
		return false
	}
	return !strings.EqualFold(OSLtoString(a), OSLtoString(b))
}

func OSLcastInt(i any) int {
	if i == nil {
		return 0
	}
	switch i := i.(type) {
	case string:
		f, _ := strconv.ParseFloat(string(i), 64)
		return int(f)
	case int:
		return i
	case float64:
		return int(i)
	case bool:
		if i {
			return 1
		}
		return 0
	case int8:
		return int(i)
	case int16:
		return int(i)
	case int32:
		return int(i)
	case int64:
		return int(i)
	case json.Number:
		f, _ := i.Float64()
		return int(f)
	default:
		panic("OSLcastInt, invalid type: " + reflect.TypeOf(i).String())
	}
}

func OSLlogValues(values ...any) {
	for _, v := range values {
		OSLlog(v)
	}
}

func OSLlog(v any) {
	if v == nil {
		fmt.Println("null")
	}
	switch v := v.(type) {
	case *SafeMap[string, any]:
		// Convert to regular map for JSON serialization
		keys := v.Keys()
		m := make(map[string]any, len(keys))
		for _, k := range keys {
			val, _ := v.Get(k)
			m[k] = val
		}
		fmt.Println(JsonStringify(m))
		return
	case *SafeSlice[any]:
		// Convert to regular slice for JSON serialization
		fmt.Println(JsonStringify(v.Values()))
		return
	case map[string]any:
		fmt.Println(JsonStringify(v))
		return
	case []any:
		fmt.Println(JsonStringify(v))
		return
	case string, int, float64, bool:
		fmt.Println(v)
		return
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			fmt.Println("null")
			return
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		fmt.Println(JsonStringify(OSLcastArray(v)))
		return
	}

	if rv.Kind() == reflect.Map {
		fmt.Println(JsonStringify(OSLcastObject(v)))
		return
	}

	fmt.Println(v)
}

func OSLisFunc(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Func
}

func OSLcallFunc(fn any, self any, params []any) any {
	if fn == nil {
		return nil
	}

	if params == nil {
		params = []any{}
	}

	if self != nil {
		params = append([]any{self}, params...)
	}

	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		panic("OSLcallFunc: invalid type: " + reflect.TypeOf(fn).String())
	}

	ft := rv.Type()
	numIn := ft.NumIn()

	isVariadic := ft.IsVariadic()

	args := make([]reflect.Value, 0, len(params))

	for i := range params {
		var pt reflect.Type

		if isVariadic && i >= numIn-1 {
			pt = ft.In(numIn - 1).Elem()
		} else {
			pt = ft.In(i)
		}

		var av reflect.Value

		if params[i] == nil {
			switch pt.Kind() {
			case reflect.Interface, reflect.Pointer, reflect.Map,
				reflect.Slice, reflect.Func, reflect.Chan:
				av = reflect.Zero(pt)
			default:
				panic("OSLcallFunc: nil is not assignable to " + pt.String())
			}
		} else {
			av = reflect.ValueOf(params[i])

			at := av.Type()

			if at.AssignableTo(pt) {
			} else if at.ConvertibleTo(pt) {
				av = av.Convert(pt)
			} else if pt.Kind() == reflect.Interface && at.Implements(pt) {
			} else {
				panic(
					"OSLcallFunc: cannot use " + at.String() +
						" as " + pt.String(),
				)
			}
		}

		args = append(args, av)
	}

	out := rv.Call(args)

	switch len(out) {
	case 0:
		return nil
	case 1:
		return out[0].Interface()
	default:
		res := make([]any, len(out))
		for i := range out {
			res[i] = out[i].Interface()
		}
		return res
	}
}

func OSLsort(arr []any) []any {
	if arr == nil {
		return nil
	}

	sort.Slice(arr, func(i, j int) bool {
		return OSLtoString(arr[i]) < OSLtoString(arr[j])
	})
	return arr
}

func OSLreplace(s string, old string, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func OSLreplaceFirst(s string, old string, new string) string {
	return strings.Replace(s, old, new, 1)
}

func OSLsortBy(arr []any, key any) []any {
	if arr == nil {
		return nil
	}

	if OSLisFunc(key) {
		sort.Slice(arr, func(i, j int) bool {
			ki := OSLcallFunc(key, nil, []any{arr[i]})
			kj := OSLcallFunc(key, nil, []any{arr[j]})

			return OSLless(ki, kj)
		})
		return arr
	}

	keyStr := OSLtoString(key)
	sort.Slice(arr, func(i, j int) bool {
		ai, ok1 := arr[i].(map[string]any)
		aj, ok2 := arr[j].(map[string]any)

		if !ok1 || !ok2 {
			return false
		}

		ki := ai[keyStr]
		kj := aj[keyStr]

		return OSLless(ki, kj)
	})

	return arr
}

func OSLless(a any, b any) bool {
	if a == b {
		return false
	}
	return OSLtoString(a) < OSLtoString(b)
}

func OSLgreater(a any, b any) bool {
	if a == b {
		return false
	}
	return OSLtoString(a) > OSLtoString(b)
}

func OSLcastNumber(n any) float64 {
	if n == nil {
		return 0
	}
	switch n := n.(type) {
	case string:
		f, _ := strconv.ParseFloat(string(n), 64)
		return f
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case bool:
		if n {
			return float64(1)
		}
		return float64(0)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return float64(n.(float64))
	}
}

func OSLcastBool(b any) bool {
	if b == nil {
		return false
	}

	switch b := b.(type) {
	case string:
		return len(b) > 0
	case int:
		return b == 1
	case bool:
		return b
	case []any:
		return len(b) > 0
	case map[string]any:
		return len(b) > 0
	default:
		v := reflect.ValueOf(b)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			return OSLcastBool(v.Elem().Interface())
		}
		panic("OSLcastBool, invalid type: " + v.Kind().String())
	}
}

func OSLcastUsable(s any) any {
	switch s := s.(type) {
	case string, int, bool, float64, map[string]any:
		return s
	case []any:
		result := make([]any, len(s))
		for i, v := range s {
			result[i] = OSLcastUsable(v)
		}
		return result
	default:
		rv := reflect.ValueOf(s)
		if rv.Kind() == reflect.Slice {
			result := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = OSLcastUsable(rv.Index(i).Interface())
			}
			return result
		}
		return fmt.Sprintf("%v", s)
	}
}

func OSLrandom[T int | float64](low, high T) T {
	if high <= low {
		return low
	}

	switch any(low).(type) {
	case int:
		return T(OSLrand.Intn(int(high-low)) + int(low))

	case float64:
		return (T(OSLrand.Float64()) * (high - low)) + low
	}

	panic("OSLrandom: unsupported type")
}

func OSLnullishCoaless(a any, b any) any {
	if a == nil {
		return b
	}
	return a
}

func OSLSplit(s string, sep string) []any {
	split := strings.Split(s, sep)
	out := make([]any, len(split))
	for i, v := range split {
		out[i] = v
	}
	return out
}

func JsonStringify(obj any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(obj); err != nil {
		return ""
	}
	return strings.TrimRight(buf.String(), "\n")
}

func JsonParse(str string) any {
	if strings.TrimSpace(str) == "" {
		return interface{}(nil)
	}

	var obj any
	decoder := json.NewDecoder(strings.NewReader(str))
	decoder.UseNumber()
	if err := decoder.Decode(&obj); err != nil {
		return interface{}(nil)
	}
	return obj
}

func JsonFormat(obj any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		return ""
	}
	return strings.TrimRight(buf.String(), "\n")
}

// Math operation wrappers for OSL behavior

func input(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func OSLgetItem(a any, b any) any {
	if a == nil {
		return nil
	}

	if sm, ok := a.(*SafeMap[string, any]); ok {
		val, _ := sm.Get(OSLtoString(b))
		return val
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		idx := OSLcastInt(b) - 1 // OSL 1-indexed
		val, ok := ss.Get(idx)
		if !ok {
			return nil
		}
		return val
	}

	if v, ok := a.(map[string]any); ok {
		return v[OSLtoString(b)]
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	key := OSLtoString(b)

	switch v.Kind() {
	case reflect.Map:
		mk := reflect.ValueOf(key)
		val := v.MapIndex(mk)
		if val.IsValid() {
			return val.Interface()
		}
	case reflect.Slice, reflect.Array:
		idx := OSLcastInt(b) - 1 // OSL 1-indexed
		if idx < 0 || idx >= v.Len() {
			return nil
		}
		return v.Index(idx).Interface()
	case reflect.Struct:
		// Try exact field name
		field := v.FieldByName(key)
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
		// Optionally: loop through fields and match lowercase names
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if strings.EqualFold(f.Name, key) && v.Field(i).CanInterface() {
				return v.Field(i).Interface()
			}
		}
	case reflect.String:
		idx := OSLcastInt(b) - 1
		s := v.String()
		if idx < 0 || idx >= len(s) {
			return ""
		}
		return string(s[idx])
	default:
		panic("OSLgetItem: invalid type (" + v.Kind().String() + ")")
	}

	return nil
}

func OSLjoin[T string | []any, T2 string | []any](a T, b T2) T {
	switch aSlice := any(a).(type) {
	case []any:
		switch bVal := any(b).(type) {
		case []any:
			return any(append(aSlice, bVal...)).(T)
		}
	}

	return any(OSLtoString(a) + OSLtoString(b)).(T)
}

func OSLadd[T float64 | int](a T, b T) T {
	return T(OSLcastNumber(a) + OSLcastNumber(b))
}

func OSLsub[T float64 | int](a T, b T) T {
	return T(OSLcastNumber(a) - OSLcastNumber(b))
}

func OSLmultiply[AT float64 | int | string, BT float64 | int](a AT, b BT) AT {
	if str, ok := any(a).(string); ok {
		n := OSLcastNumber(b)
		var out string
		if n < 0 {
			out = ""
		}
		out = strings.Repeat(str, int(n))
		if n < 0 {
			out = ""
		}
		return any(out).(AT)
	}

	return any(OSLcastNumber(a) * OSLcastNumber(b)).(AT)
}

func OSLdivide[T float64 | int](a T, b T) float64 {
	return float64(OSLcastNumber(a) / OSLcastNumber(b))
}

func OSLmod[T float64 | int](a T, b T) T {
	return T(math.Mod(OSLcastNumber(a), OSLcastNumber(b)))
}

func OSLmin[T float64 | int](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

func OSLmax[T float64 | int](a T, b T) T {
	if a > b {
		return a
	}
	return b
}

func OSLround(n any) int {
	if n == nil {
		return 0
	}
	switch n := n.(type) {
	case int:
		return n
	case float64:
		return int(n + 0.5)
	default:
		panic("OSLround, invalid type: " + reflect.TypeOf(n).String())
	}
}

func OSLceil(n any) float64 {
	switch n := n.(type) {
	case int:
		return float64(n)
	case float64:
		return math.Ceil(n)
	default:
		panic("OSLceil, invalid type: " + reflect.TypeOf(n).String())
	}
}

func OSLfloor(n any) float64 {
	switch n := n.(type) {
	case int:
		return float64(n)
	case float64:
		return math.Floor(n)
	default:
		panic("OSLfloor, invalid type: " + reflect.TypeOf(n).String())
	}
}

func OSLtrim[S string | []any, F int | float64, T int | float64](s S, from F, to T) S {
	var items []any
	isArr := false

	if arr, ok := any(s).([]any); ok {
		items = arr
		isArr = true
	} else {
		items = make([]any, 0)
		for _, r := range []rune(OSLtoString(s)) {
			items = append(items, string(r))
		}
	}

	n := len(items)
	start := int(from) - 1
	end := int(to)

	if start < 0 {
		start = 0
	} else if start > n {
		start = n
	}
	if end < 0 {
		end = n + end + 1
	}
	if end > n {
		end = n
	} else if end < 0 {
		end = 0
	}
	if start > end {
		start, end = end, start
	}

	if isArr {
		return any(items[start:end]).(S)
	}
	result := make([]rune, len(items[start:end]))
	for i, v := range items[start:end] {
		result[i] = []rune(v.(string))[0]
	}
	return any(string(result)).(S)
}

func OSLwait(seconds float64) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

func OSLslice(s any, start int, end int) []any {
	arr := OSLcastArray(s)
	n := len(arr)

	start = start - 1
	if start < 0 {
		start = 0
	} else if start > n {
		start = n
	}

	if end < 0 {
		end = n + end + 1
	}
	if end > n {
		end = n
	} else if end < 0 {
		end = 0
	}

	if start > end {
		start, end = end, start
	}

	return arr[start:end]
}

func OSLpadStart(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(pad, length-len(s)) + s
}

func OSLpadEnd(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(pad, length-len(s))
}

func OSLtypeof(s any) string {
	switch s.(type) {
	case string:
		return "string"
	case int:
		return "int"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return "any"
	}
}

func OSLKeyIn(b any, a any) bool {
	if a == nil {
		return false
	}

	key := OSLtoString(b)
	if sm, ok := a.(*SafeMap[string, any]); ok {
		_, exists := sm.Get(key)
		return exists
	}

	switch a := a.(type) {
	case map[string]any:
		_, ok := a[key]
		return ok
	case []any:
		for _, v := range a {
			if OSLtoString(v) == key {
				return true
			}
		}
		return false
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		mapKeyType := v.Type().Key()
		mk := reflect.ValueOf(key)
		if !mk.Type().AssignableTo(mapKeyType) {
			if mapKeyType.Kind() == reflect.String {
				mk = reflect.ValueOf(key)
			} else {
				return false
			}
		}
		val := v.MapIndex(mk)
		return val.IsValid()

	case reflect.Slice, reflect.Array:
		idx := OSLcastInt(b) - 1
		return idx >= 0 && idx < v.Len()

	case reflect.String:
		idx := OSLcastInt(b) - 1
		return idx >= 0 && idx < len(v.String())

	case reflect.Struct:
		if field := v.FieldByName(key); field.IsValid() {
			return true
		}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if strings.EqualFold(f.Name, key) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

func OSLdelete(a any, b any) any {
	if a == nil {
		return nil
	}

	if sm, ok := a.(*SafeMap[string, any]); ok {
		sm.Delete(OSLtoString(b))
		return a
	}

	switch a := a.(type) {
	case map[string]any:
		delete(a, OSLtoString(b))
		return a
	case []any:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= len(a) {
			return a
		}
		return append(a[:idx], a[idx+1:]...)
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	key := OSLtoString(b)

	switch v.Kind() {
	case reflect.Map:
		mk := reflect.ValueOf(key)
		if mk.Type().AssignableTo(v.Type().Key()) {
			v.SetMapIndex(mk, reflect.Value{})
		}
		return v.Interface()

	case reflect.Slice:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= v.Len() {
			return v.Interface()
		}
		newSlice := reflect.AppendSlice(v.Slice(0, idx), v.Slice(idx+1, v.Len()))
		return newSlice.Interface()

	default:
		return a
	}
}

func OSLsetItem(a any, b any, value any) bool {
	if a == nil {
		return false
	}

	if sm, ok := a.(*SafeMap[string, any]); ok {
		sm.Set(OSLtoString(b), value)
		return true
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		idx := OSLcastInt(b) - 1
		return ss.Set(idx, value)
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	key := OSLtoString(b)

	switch v.Kind() {
	case reflect.Map:
		mk := reflect.ValueOf(key)
		if !mk.IsValid() {
			return false
		}

		var mv reflect.Value
		if value == nil {
			mv = reflect.Zero(v.Type().Elem())
		} else {
			mv = reflect.ValueOf(value)
		}

		if mk.Type().AssignableTo(v.Type().Key()) && mv.Type().AssignableTo(v.Type().Elem()) {
			v.SetMapIndex(mk, mv)
			return true
		}
		return false

	case reflect.Slice:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= v.Len() {
			return false
		}
		elem := reflect.ValueOf(value)
		if elem.Type().AssignableTo(v.Index(idx).Type()) {
			v.Index(idx).Set(elem)
			return true
		}
		return false

	case reflect.Struct:
		field := v.FieldByName(key)
		if !field.IsValid() {
			return false
		}

		var val reflect.Value
		if value == nil {
			val = reflect.Zero(field.Type())
		} else {
			val = reflect.ValueOf(value)
		}

		return setFieldUnsafe(field, val)
	}

	return false
}

func setFieldUnsafe(field reflect.Value, val reflect.Value) bool {
	if !field.CanAddr() {
		return false
	}

	if !val.Type().AssignableTo(field.Type()) {
		if val.Type().ConvertibleTo(field.Type()) {
			val = val.Convert(field.Type())
		} else {
			return false
		}
	}

	ptr := unsafe.Pointer(field.UnsafeAddr())
	reflect.NewAt(field.Type(), ptr).Elem().Set(val)
	return true
}

func OSLarrayJoin(a any, b any) string {
	var out strings.Builder
	sep := OSLtoString(b)
	arr := OSLcastArray(a)

	for _, v := range arr {
		out.WriteString(OSLtoString(v) + sep)
	}

	return strings.TrimSuffix(out.String(), sep)
}

func OSLgetKeys(a any) []any {
	if sm, ok := a.(*SafeMap[string, any]); ok {
		keys := sm.Keys()
		result := make([]any, len(keys))
		for i, k := range keys {
			result[i] = k
		}
		return result
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		length := ss.Len()
		keys := make([]any, length)
		for i := 0; i < length; i++ {
			keys[i] = i + 1 // OSL is 1-indexed
		}
		return keys
	}

	switch a := a.(type) {
	case map[string]any:
		keys := make([]any, len(a))
		i := 0
		for k := range a {
			keys[i] = k
			i++
		}
		return keys
	case []any:
		keys := make([]any, len(a))
		for i := range a {
			keys[i] = i + 1 // OSL is 1-indexed
		}
		return keys
	default:
		return []any{}
	}
}

func OSLgetValues(a any) []any {
	if sm, ok := a.(*SafeMap[string, any]); ok {
		values := sm.Values()
		result := make([]any, len(values))
		copy(result, values)
		return result
	}

	switch a := a.(type) {
	case map[string]any:
		values := make([]any, len(a))
		i := 0
		for _, v := range a {
			values[i] = v
			i++
		}
		return values
	case []any:
		values := make([]any, len(a))
		i := 0
		for _, v := range a {
			values[i] = v
			i++
		}
		return values
	default:
		return []any{}
	}
}

func OSLcontains(a any, b any) bool {
	if sm, ok := a.(*SafeMap[string, any]); ok {
		_, exists := sm.Get(OSLtoString(b))
		return exists
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		// For arrays, check if value exists
		values := ss.Values()
		for _, v := range values {
			if OSLtoString(v) == OSLtoString(b) {
				return true
			}
		}
		return false
	}

	switch a := a.(type) {
	case map[string]any:
		_, ok := a[OSLtoString(b)]
		return ok
	case []any:
		for _, v := range a {
			if OSLtoString(v) == OSLtoString(b) {
				return true
			}
		}
		return false
	case string:
		return strings.Contains(a, OSLtoString(b))
	default:
		return false
	}
}

func OSLappend(a *[]any, b any) []any {
	*a = append(*a, b)
	return *a
}

func OSLpop(a *[]any) any {
	if len(*a) == 0 {
		return nil
	}
	last := (*a)[len(*a)-1]
	*a = (*a)[:len(*a)-1]
	return last
}

func OSLshift(a *[]any) any {
	if len(*a) == 0 {
		return nil
	}
	first := (*a)[0]
	*a = append([]any{}, (*a)[1:]...)
	return first
}

func OSLprepend(a *[]any, b any) []any {
	*a = append([]any{b}, *a...)
	return *a
}

func OSLclone(a any) any {
	switch a := a.(type) {
	case map[string]any:
		b := make(map[string]any, len(a))
		for k, v := range a {
			b[k] = OSLclone(v)
		}
		return b
	case []any:
		b := make([]any, len(a))
		for i, v := range a {
			b[i] = OSLclone(v)
		}
		return b
	default:
		return a
	}
}

// worker handling

var OSLself any = nil

func OSLworker(props map[string]any) map[string]any {
	props["createdTime"] = time.Now()
	props["processTime"] = 0
	props["alive"] = true
	props["kill"] = func() {
		props["alive"] = false
	}
	go (func() {
		OSLself = props
		OSLcallFunc(props["oncreate"], props, nil)
		for {
			startTime := time.Now()
			OSLself = props
			OSLcallFunc(props["onframe"], props, nil)
			props["processTime"] = OSLcastNumber(props["processTime"]) + time.Since(startTime).Seconds()
			if props["alive"] != true {
				props["alive"] = false
				OSLself = props
				OSLcallFunc(props["onkill"], props, nil)
				break
			}
		}
	})()
	return props
}

type SafeMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewSafeMap[K comparable, V any](defaults map[K]V) *SafeMap[K, V] {
	sm := &SafeMap[K, V]{
		data: make(map[K]V, len(defaults)),
	}
	for k, v := range defaults {
		sm.data[k] = v
	}
	return sm
}

func (m *SafeMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value // regular map syntax here
}

func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key] // regular map syntax here
	return value, ok
}

func (m *SafeMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *SafeMap[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *SafeMap[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// SafeSlice is a thread-safe slice for global arrays
type SafeSlice[V any] struct {
	mu   sync.RWMutex
	data []V
}

func NewSafeSlice[V any](defaults []V) *SafeSlice[V] {
	ss := &SafeSlice[V]{
		data: make([]V, len(defaults)),
	}
	copy(ss.data, defaults)
	return ss
}

func (s *SafeSlice[V]) Append(value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, value)
}

func (s *SafeSlice[V]) Get(index int) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index < 0 || index >= len(s.data) {
		var zero V
		return zero, false
	}
	return s.data[index], true
}

func (s *SafeSlice[V]) Set(index int, value V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.data) {
		return false
	}
	s.data[index] = value
	return true
}

func (s *SafeSlice[V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *SafeSlice[V]) Values() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]V, len(s.data))
	copy(values, s.data)
	return values
}

// Keyboard methods (stub implementations)
// Note: These are defined as methods on a custom string type
type OSLString string

func (s OSLString) onKeyDown() bool {
	return false
}

func (s OSLString) isKeyDown() bool {
	return false
}

func (s OSLString) toNum() float64 {
	return OSLcastNumber(string(s))
}

func atob(encoded string) string {
	data, err := OSLio.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded)))
	if err != nil {
		return ""
	}
	return string(data)
}

func btoa(data string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return encoded
}



// name: img
// description: Memory-efficient image utilities following Go idioms
// author: Mist
// requires: bytes as OSL_bytes, image as OSL_image, image/png, image/jpeg, golang.org/x/image/draw as OSL_draw, github.com/nfnt/resize as OSL_img_resize, os, io, runtime, github.com/rwcarlsen/goexif/exif as OSL_exif, image/color as OSL_color

type IMG struct{}

type OSL_img_Image struct {
	im     *OSL_image.RGBA
	closed bool
}

/* -------------------- helpers -------------------- */

func OSL_allocRGBA(w, h int) *OSL_image.RGBA {
	return OSL_image.NewRGBA(OSL_image.Rect(0, 0, w, h))
}

func OSL_toRGBA(src OSL_image.Image) *OSL_image.RGBA {
	if src == nil {
		return nil
	}

	if im, ok := src.(*OSL_image.RGBA); ok {
		dst := OSL_allocRGBA(im.Bounds().Dx(), im.Bounds().Dy())
		copy(dst.Pix, im.Pix)
		return dst
	}

	b := src.Bounds()
	dst := OSL_allocRGBA(b.Dx(), b.Dy())
	OSL_draw.Draw(dst, dst.Bounds(), src, b.Min, OSL_draw.Src)
	return dst
}

func OSL_newImage(im *OSL_image.RGBA) *OSL_img_Image {
	if im == nil {
		return nil
	}
	return &OSL_img_Image{im: im}
}

/* -------------------- lifecycle -------------------- */

func (i *OSL_img_Image) Close() {
	if i == nil || i.closed {
		return
	}
	i.closed = true
	i.im = nil
}

func (i *OSL_img_Image) RGBA() *OSL_image.RGBA {
	if i == nil || i.closed {
		return nil
	}
	return i.im
}

/* -------------------- metadata -------------------- */

func (i *OSL_img_Image) Width() int {
	if i == nil || i.closed || i.im == nil {
		return 0
	}
	return i.im.Bounds().Dx()
}

func (i *OSL_img_Image) Height() int {
	if i == nil || i.closed || i.im == nil {
		return 0
	}
	return i.im.Bounds().Dy()
}

func (i *OSL_img_Image) Size() map[string]any {
	if i == nil || i.closed || i.im == nil {
		return map[string]any{}
	}
	b := i.im.Bounds()
	return map[string]any{"w": b.Dx(), "h": b.Dy()}
}

/* -------------------- decode helpers -------------------- */

func (IMG) Open(path string) *OSL_img_Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	im, _, err := OSL_image.Decode(f)
	if err != nil {
		return nil
	}
	return OSL_newImage(OSL_toRGBA(im))
}

func (IMG) Decode(r OSLio.Reader) *OSL_img_Image {
	im, _, err := OSL_image.Decode(r)
	if err != nil {
		return nil
	}
	return OSL_newImage(OSL_toRGBA(im))
}

func (IMG) DecodeBytes(data []byte) *OSL_img_Image {
	return img.Decode(OSL_bytes.NewReader(data))
}

func (IMG) OpenSize(path string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	cfg, _, err := OSL_image.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func (IMG) DecodeSize(r OSLio.Reader) (int, int) {
	cfg, _, err := OSL_image.DecodeConfig(r)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

/* -------------------- encode -------------------- */

func (IMG) EncodePNG(w OSLio.Writer, i *OSL_img_Image) bool {
	return i != nil && !i.closed && png.Encode(w, i.im) == nil
}

func (IMG) EncodeJPEG(w OSLio.Writer, i *OSL_img_Image, q int) bool {
	if i == nil || i.closed {
		return false
	}
	if q < 1 {
		q = 1
	} else if q > 100 {
		q = 100
	}
	return jpeg.Encode(w, i.im, &jpeg.Options{Quality: q}) == nil
}

func (IMG) EncodePNGBytes(i *OSL_img_Image) []byte {
	if i == nil || i.closed {
		return nil
	}
	var buf OSL_bytes.Buffer
	if png.Encode(&buf, i.im) != nil {
		return nil
	}
	return buf.Bytes()
}

func (IMG) EncodeJPEGBytes(i *OSL_img_Image, q int) []byte {
	if i == nil || i.closed {
		return nil
	}
	if q < 1 {
		q = 1
	} else if q > 100 {
		q = 100
	}
	var buf OSL_bytes.Buffer
	if jpeg.Encode(&buf, i.im, &jpeg.Options{Quality: q}) != nil {
		return nil
	}
	return buf.Bytes()
}

/* -------------------- creation -------------------- */

func (IMG) New(w, h int) *OSL_img_Image {
	if w <= 0 || h <= 0 {
		return nil
	}
	return OSL_newImage(OSL_allocRGBA(w, h))
}

func (IMG) Clone(i *OSL_img_Image) *OSL_img_Image {
	if i == nil || i.closed {
		return nil
	}
	b := i.im.Bounds()
	dst := OSL_allocRGBA(b.Dx(), b.Dy())
	copy(dst.Pix, i.im.Pix)
	return OSL_newImage(dst)
}

/* -------------------- resize helpers -------------------- */

func (IMG) Resize(i *OSL_img_Image, w, h int) *OSL_img_Image {
	if i == nil || i.closed || (w == 0 && h == 0) || w < 0 || h < 0 {
		return nil
	}
	r := OSL_img_resize.Resize(uint(w), uint(h), i.im, OSL_img_resize.Lanczos3)
	return OSL_newImage(OSL_toRGBA(r))
}

func (IMG) ResizeFast(i *OSL_img_Image, w, h int) *OSL_img_Image {
	if i == nil || i.closed || (w == 0 && h == 0) || w < 0 || h < 0 {
		return nil
	}
	r := OSL_img_resize.Resize(uint(w), uint(h), i.im, OSL_img_resize.Bilinear)
	return OSL_newImage(OSL_toRGBA(r))
}

func (IMG) ResizeWidth(i *OSL_img_Image, w int) *OSL_img_Image {
	return img.Resize(i, w, 0)
}

func (IMG) ResizeHeight(i *OSL_img_Image, h int) *OSL_img_Image {
	return img.Resize(i, 0, h)
}

func (IMG) ResizeFit(i *OSL_img_Image, maxW, maxH int) *OSL_img_Image {
	if i == nil || i.closed {
		return nil
	}
	sw, sh := i.Width(), i.Height()
	rw := float64(maxW) / float64(sw)
	rh := float64(maxH) / float64(sh)
	scale := math.Min(rw, rh)
	return img.Resize(i, int(float64(sw)*scale), int(float64(sh)*scale))
}

/* -------------------- draw / composite -------------------- */

func (IMG) Draw(dst, src *OSL_img_Image, x, y int) bool {
	if dst == nil || src == nil || dst.closed || src.closed {
		return false
	}
	r := OSL_image.Rect(x, y, x+src.Width(), y+src.Height())
	OSL_draw.Draw(dst.im, r, src.im, OSL_image.Point{}, OSL_draw.Src)
	return true
}

func (IMG) DrawOver(dst, src *OSL_img_Image, x, y int) bool {
	if dst == nil || src == nil || dst.closed || src.closed {
		return false
	}
	r := OSL_image.Rect(x, y, x+src.Width(), y+src.Height())
	OSL_draw.Draw(dst.im, r, src.im, OSL_image.Point{}, OSL_draw.Over)
	return true
}

/* -------------------- rotation -------------------- */

func (IMG) Rotate(i *OSL_img_Image, angle float64) *OSL_img_Image {
	if i == nil || i.closed {
		return nil
	}
	a := int(angle) % 360
	if a < 0 {
		a += 360
	}
	if a%90 == 0 {
		return img.rotate90(i.im, a)
	}
	return img.rotateAny(i.im, angle)
}

func (IMG) rotate90(src *OSL_image.RGBA, a int) *OSL_img_Image {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	if a == 0 {
		dst := OSL_allocRGBA(sw, sh)
		copy(dst.Pix, src.Pix)
		return OSL_newImage(dst)
	}

	var dst *OSL_image.RGBA
	if a == 180 {
		dst = OSL_allocRGBA(sw, sh)
	} else {
		dst = OSL_allocRGBA(sh, sw)
	}

	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			c := src.At(x, y)
			switch a {
			case 90:
				dst.Set(sh-1-y, x, c)
			case 180:
				dst.Set(sw-1-x, sh-1-y, c)
			case 270:
				dst.Set(y, sw-1-x, c)
			}
		}
	}
	return OSL_newImage(dst)
}

func (IMG) rotateAny(src *OSL_image.RGBA, angle float64) *OSL_img_Image {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	rad := angle * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)

	nw := int(math.Abs(float64(sw)*cos) + math.Abs(float64(sh)*sin))
	nh := int(math.Abs(float64(sw)*sin) + math.Abs(float64(sh)*cos))

	dst := OSL_allocRGBA(nw, nh)

	cx, cy := float64(sw)/2, float64(sh)/2
	ncx, ncy := float64(nw)/2, float64(nh)/2

	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			tx := float64(x) - ncx
			ty := float64(y) - ncy

			sx := tx*cos + ty*sin + cx
			sy := -tx*sin + ty*cos + cy

			ix := int(math.Round(sx))
			iy := int(math.Round(sy))

			if ix >= 0 && iy >= 0 && ix < sw && iy < sh {
				dst.Set(x, y, src.At(ix, iy))
			}
		}
	}
	return OSL_newImage(dst)
}

/* -------------------- exif orientation -------------------- */

func (IMG) NormalizeOrientation(i *OSL_img_Image, r OSLio.Reader) *OSL_img_Image {
	if i == nil || i.closed {
		return nil
	}

	ex, err := OSL_exif.Decode(r)
	if err != nil {
		return i
	}

	tag, err := ex.Get(OSL_exif.Orientation)
	if err != nil {
		return i
	}

	o, _ := tag.Int(0)

	switch o {
	case 3:
		return img.Rotate(i, 180)
	case 6:
		return img.Rotate(i, 90)
	case 8:
		return img.Rotate(i, 270)
	default:
		return i
	}
}

/* -------------------- fill / color helpers -------------------- */

func (IMG) Fill(i *OSL_img_Image, r, g, b, a uint8) bool {
	if i == nil || i.closed {
		return false
	}
	p := i.im.Pix
	for j := 0; j < len(p); j += 4 {
		p[j+0] = r
		p[j+1] = g
		p[j+2] = b
		p[j+3] = a
	}
	return true
}

func RGB(r, g, b uint8) OSL_color.RGBA {
	return OSL_color.RGBA{R: r, G: g, B: b, A: 255}
}

func RGBA(r, g, b, a uint8) OSL_color.RGBA {
	return OSL_color.RGBA{R: r, G: g, B: b, A: a}
}

/* -------------------- saving helpers -------------------- */

func (IMG) SavePNG(i *OSL_img_Image, path string) bool {
	if i == nil || i.closed || i.im == nil {
		return false
	}

	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	return png.Encode(f, i.im) == nil
}

func (IMG) SaveJPEG(i *OSL_img_Image, path string, quality int) bool {
	if i == nil || i.closed || i.im == nil {
		return false
	}

	if quality < 1 {
		quality = 1
	} else if quality > 100 {
		quality = 100
	}

	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	return jpeg.Encode(f, i.im, &jpeg.Options{Quality: quality}) == nil
}

var img = IMG{}
// name: requests
// description: HTTP utilities
// author: Mist
// requires: net/http, encoding/json, io

type HTTP struct {
	Client *http.Client
}

func extractHeadersAndBody(data map[string]any) (headers map[string]string, body OSLio.Reader) {
	headers = make(map[string]string)
	if data != nil {
		if raw, ok := data["body"]; ok {
			switch v := raw.(type) {
			case string:
				body = bytes.NewReader([]byte(v))
			case []byte:
				body = bytes.NewReader(v)
			case map[string]any:
				body = bytes.NewReader([]byte(JsonStringify(v)))
				headers["Content-Type"] = "application/json"
			default:
				buf, _ := json.Marshal(v)
				body = bytes.NewReader(buf)
				headers["Content-Type"] = "application/json"
			}
		}

		for k, v := range data {
			if k == "body" {
				continue
			}
			headers[k] = OSLtoString(v)
		}
	}
	return headers, body
}

func (h *HTTP) doRequest(method, url string, data map[string]any) map[string]any {
	headers, body := extractHeadersAndBody(data)

	out := make(map[string]any)
	out["headers"] = nil
	out["body"] = nil
	out["raw"] = nil
	out["status"] = 0
	out["success"] = false

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return out
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return out
	}
	defer resp.Body.Close()

	respHeaders := make(map[string]any)
	for k, v := range resp.Header {
		if len(v) == 1 {
			respHeaders[k] = v[0]
		} else {
			respHeaders[k] = v
		}
	}

	out["status"] = resp.StatusCode
	out["headers"] = respHeaders
	out["raw"] = resp
	out["success"] = true

	respBody, err := OSLio.ReadAll(resp.Body)
	if err != nil {
		return out
	}

	out["body"] = respBody

	return out
}

func (h *HTTP) Get(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodGet, OSLtoString(url), m)
}

func (h *HTTP) Post(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPost, OSLtoString(url), data)
}

func (h *HTTP) Put(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPut, OSLtoString(url), data)
}

func (h *HTTP) Patch(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPatch, OSLtoString(url), data)
}

func (h *HTTP) Delete(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodDelete, OSLtoString(url), m)
}

func (h *HTTP) Options(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodOptions, OSLtoString(url), m)
}

func (h *HTTP) Head(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	out := map[string]any{"success": false}
	headers, _ := extractHeadersAndBody(m)
	req, err := http.NewRequest(http.MethodHead, OSLtoString(url), nil)
	if err != nil {
		return out
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := h.Client.Do(req)
	if err != nil {
		return out
	}
	out["status"] = resp.StatusCode
	defer resp.Body.Close()
	out["success"] = true
	return out
}

var requests = &HTTP{Client: http.DefaultClient}
// name: fs
// description: File system utilities
// author: Mist
// requires: os, path/filepath

type FS struct{}

func (FS) ReadFile(path any) string {
	data, err := os.ReadFile(OSLtoString(path))
	if err != nil {
		return ""
	}
	return string(data)
}

func (FS) ReadFileBytes(path any) []byte {
	data, err := os.ReadFile(OSLtoString(path))
	if err != nil {
		return []byte{}
	}
	return data
}

func (FS) WriteFile(path any, data any) bool {
	err := os.WriteFile(OSLtoString(path), []byte(OSLtoString(data)), 0644)
	return err == nil
}

func (FS) Rename(oldPath any, newPath any) bool {
	err := os.Rename(OSLtoString(oldPath), OSLtoString(newPath))
	return err == nil
}

func (FS) Exists(path any) bool {
	_, err := os.Stat(OSLtoString(path))
	return err == nil
}

func (FS) Remove(path any) bool {
	pathStr, ok := path.(string)
	if !ok || pathStr == "" {
		return false
	}

	info, err := os.Stat(pathStr)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return os.RemoveAll(pathStr) == nil
	}

	return os.Remove(pathStr) == nil
}

func (FS) Mkdir(path any) bool {
	err := os.Mkdir(OSLtoString(path), 0755)
	return err == nil
}

func (FS) MkdirAll(path any) bool {
	err := os.MkdirAll(OSLtoString(path), 0755)
	return err == nil
}

func (FS) CopyDir(srcPath any, dstPath any) bool {
	src := OSLtoString(srcPath)
	dst := OSLtoString(dstPath)

	entries, err := os.ReadDir(src)
	if err != nil {
		return false
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return false
	}

	for _, entry := range entries {
		srcFile := filepath.Join(src, entry.Name())
		dstFile := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			ok := (FS{}).CopyDir(srcFile, dstFile)
			if !ok {
				return false
			}
			continue
		}

		in, err := os.Open(srcFile)
		if err != nil {
			return false
		}

		out, err := os.Create(dstFile)
		if err != nil {
			in.Close()
			return false
		}

		if _, err := OSLio.Copy(out, in); err != nil {
			in.Close()
			out.Close()
			return false
		}

		in.Close()
		out.Close()

		if info, err := os.Stat(srcFile); err == nil {
			_ = os.Chmod(dstFile, info.Mode())
		}
	}

	return true
}

func (FS) ReadDir(path any) []any {
	files, err := os.ReadDir(OSLtoString(path))
	if err != nil {
		return []any{}
	}
	names := make([]any, len(files))
	for i, f := range files {
		names[i] = f.Name()
	}
	return names
}

func (FS) ReadDirAll(path any) []map[string]any {
	dir := OSLtoString(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []map[string]any{}
	}

	filesOut := make([]map[string]any, len(entries))
	for i, f := range entries {
		filesOut[i] = map[string]any{
			"name":  f.Name(),
			"ext":   filepath.Ext(f.Name()),
			"path":  filepath.Join(dir, f.Name()),
			"isDir": f.IsDir(),
			"type":  f.Type(),
		}
	}

	return filesOut
}

func (FS) WalkDir(path any, fn func(path string, file map[string]any, control map[string]any)) {
	dir := OSLtoString(path)
	filepath.WalkDir(dir, func(p string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		fileData := map[string]any{
			"name":    entry.Name(),
			"ext":     filepath.Ext(entry.Name()),
			"path":    p,
			"isDir":   entry.IsDir(),
			"size":    info.Size(),
			"mode":    info.Mode(),
			"modTime": info.ModTime(),
			"sys":     info.Sys(),
			"type":    entry.Type(),
		}

		control := map[string]any{
			"skip": false,
		}
		fn(p, fileData, control)
		if control["skip"] == true {
			return filepath.SkipDir
		}
		return nil
	})
}

func (FS) IsDir(path any) bool {
	info, err := os.Stat(OSLtoString(path))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (FS) Getwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func (FS) Chdir(path any) bool {
	err := os.Chdir(OSLtoString(path))
	return err == nil
}

func (FS) JoinPath(path ...any) string {
	stringPath := make([]string, len(path))
	for i, p := range path {
		stringPath[i] = OSLtoString(p)
	}
	return filepath.Join(stringPath...)
}

func (FS) GetBase(path any) string {
	return filepath.Base(OSLtoString(path))
}

func (FS) GetDir(path any) string {
	return filepath.Dir(OSLtoString(path))
}

func (FS) GetExt(path any) string {
	return filepath.Ext(OSLtoString(path))
}

func (FS) GetParts(path any) []any {
	stringPath := OSLtoString(path)
	return []any{filepath.Base(stringPath), filepath.Dir(stringPath), filepath.Ext(stringPath)}
}

func (FS) GetSize(path any) float64 {
	info, err := os.Stat(OSLtoString(path))
	if err != nil {
		return 0
	}
	return float64(info.Size())
}

func (FS) GetModTime(path any) float64 {
	info, err := os.Stat(OSLtoString(path))
	if err != nil {
		return 0.0
	}
	return float64(info.ModTime().UnixMilli())
}

func (FS) GetStat(path any) map[string]any {
	info, err := os.Stat(OSLtoString(path))
	if err != nil {
		return map[string]any{"success": false}
	}
	return map[string]any{
		"success": true,
		"name":    filepath.Base(info.Name()),
		"ext":     filepath.Ext(info.Name()),
		"path":    info.Name(),
		"isDir":   info.IsDir(),
		"size":    info.Size(),
		"mode":    info.Mode(),
		"modTime": info.ModTime().UnixMicro(),
		"sys":     info.Sys(),
	}
}

func (FS) EvalSymlinks(path any) string {
	pathStr := OSLtoString(path)
	absPath, err := filepath.EvalSymlinks(pathStr)
	if err != nil {
		return ""
	}
	return absPath
}

// Global instance
var fs = FS{}
func abortUnauthorized(c any) {
	var path string
		path = OSLtoString(OSLgetItem(OSLgetItem(OSLgetItem(c, "Request"), "URL"), "Path"))
		if strings.HasPrefix(path, "/api") {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authenticated",
				})
		} else {
				c.Redirect(302, "/auth")
		}
		c.Abort()
}
func requireSession(c any) {
	var sessionId string
	var userId string
	var data string
	var username string
	var profile map[string]any
		sessionId = OSLtoString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
		if OSLcastBool(OSLequal(sessionId, "")) {
				sessionId = c.GetHeader("sessionId")
				if OSLcastBool(OSLequal(sessionId, "")) {
						abortUnauthorized(c)
						return
				}
		}
		if (OSLcontains(sessions, sessionId) != true) {
				abortUnauthorized(c)
				return
		}
		userId = OSLtoString(OSLgetItem(sessions, sessionId))
		if (OSLcontains(userData, userId) != true) {
				if OSLcastBool(fs.Exists((("db/" + OSLtoString(userId)) + "/user.json"))) {
						data = OSLtoString(fs.ReadFile((("db/" + OSLtoString(userId)) + "/user.json")))
						OSLsetItem(userData, userId, JsonParse(data))
				}
		}
		c.Set("userId", userId)
		username = ""
		if OSLcontains(userIdToUsername, userId) {
				username = OSLtoString(OSLgetItem(userIdToUsername, userId))
		} else if OSLcontains(userData, userId) {
				profile = OSLcastObject(OSLgetItem(userData, userId))
				if OSLcastBool(OSLnotEqual(OSLgetItem(profile, "username"), nil)) {
						username = OSLtoString(OSLgetItem(profile, "username"))
				}
		}
		c.Set("username", username)
		c.Next()
}

func homePage(c any) {
	var userId string
	var username string
	var profileReq map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		username = OSLtoString(c.MustGet("username"))
		if (OSLcontains(userData, userId) != true) {
				profileReq = OSLcastObject(writeProfile(userId, username))
				if (OSLgetItem(profileReq, "ok") != true) {
						c.String(401, OSLtoString(OSLgetItem(profileReq, "error")))
						return
				}
		}
		c.HTML(200, "index.html", map[string]any{
				"Username": username,
				"UserId": userId,
				"Subscription": OSLgetItem(OSLgetItem(userData, userId), "subscription"),
		})
}
func authPage(c any) {
		c.HTML(200, "auth.html", map[string]any{
				"AuthKey": authKey,
		})
}

func randomString(length int) string{
	var chars string
	var result string
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result = ""
		for i := 1; i <= OSLround(length); i++ {
				result = (OSLtoString(result) + OSLtoString(OSLgetItem(chars, OSLrandom(1, OSLlen(chars)))))
		}
		return result
}
func calculateFileHash(path string) string{
	var hash string
		hash = ""
		f, err := os.Open(path)
		if OSLcastBool(OSLequal(err, nil)) {
				h := md5.New()
				if _, err := OSLio.Copy(h, f); err == nil { hash = hex.EncodeToString(h.Sum(nil)) }
				f.Close()
		}
		return hash
}
func noCORS(c any) {
	var origin string
		origin = OSLtoString(c.GetHeader("Origin"))
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "sessionId")
		if OSLcastBool(OSLnotEqual(origin, "")) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Next()
				return
		}
		c.Header("Access-Control-Allow-Origin", "*")
		if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "OPTIONS")) {
				c.AbortWithStatus(204)
				return
		}
		c.Next()
}
func loadConfig() {
	var config map[string]any
		config = JsonParse(fs.ReadFile("./config.json")).(map[string]any)
		authKey = OSLtoString(OSLgetItem(config, "authKey"))
		useSubscriptions = OSLcastBool(OSLgetItem(config, "useSubscriptions"))
		subscriptionSizes = OSLgetItem(config, "subscriptionSizes").(map[string]any)
		quotas = OSLgetItem(config, "quotas").(map[string]any)
		if OSLcastBool(OSLnotEqual(OSLgetItem(config, "downscaleWhen"), nil)) {
				downscaleWhen = OSLgetItem(config, "downscaleWhen").(map[string]any)
		}
}
func getAble(userId string) map[string]any{
	var profile map[string]any
	var data string
	var subscription string
	var quota float64
	var maybeQuota any
	var images []any
		profile = OSLcastObject(OSLgetItem(userData, userId))
		if OSLcastBool(OSLequal(profile, nil)) {
				if OSLcastBool(fs.Exists((("db/" + OSLtoString(userId)) + "/user.json"))) {
						data = OSLtoString(fs.ReadFile((("db/" + OSLtoString(userId)) + "/user.json")))
						profile = JsonParse(data).(map[string]any)
				} else {
						return map[string]any{
								"canAccess": false,
								"maxUpload": "0",
						}
				}
		}
		subscription = strings.ToLower(OSLtoString(OSLgetItem(profile, "subscription")))
		quota = 0
		maybeQuota = OSLgetItem(quotas, userId)
		if OSLcastBool(OSLnotEqual(maybeQuota, nil)) {
				quota = OSLcastNumber(maybeQuota)
		} else if useSubscriptions {
				quota = OSLcastNumber(OSLgetItem(subscriptionSizes, subscription))
		}
		images = readUserImages(userId)
		return map[string]any{
				"canAccess": OSLcastNumber(quota) > OSLcastNumber(0),
				"storageQuota": quota,
				"hasImages": OSLcastNumber(OSLlen(images)) > OSLcastNumber(0),
		}
}
func getProfile(username string) map[string]any{
	var resp_profile map[string]any
	var profile map[string]any
		resp_profile = OSLcastObject(requests.Get(("https://api.rotur.dev/profile?include_posts=0&name=" + OSLtoString(username))))
		if (OSLgetItem(resp_profile, "success") != true) {
				return map[string]any{
						"ok": false,
						"error": "failed to fetch profile",
				}
		}
		profile = OSLcastObject(JsonParse(OSLtoString(OSLgetItem(resp_profile, "body"))))
		if OSLcastBool(OSLnotEqual(OSLgetItem(profile, "error"), nil)) {
				return map[string]any{
						"ok": false,
						"error": OSLgetItem(profile, "error"),
				}
		}
		return map[string]any{
				"ok": true,
				"profile": profile,
		}
}
func writeProfile(userId string, username string) map[string]any{
	var profileReq map[string]any
		profileReq = getProfile(username)
		if (OSLgetItem(profileReq, "ok") != true) {
				return profileReq
		}
		OSLsetItem(userData, userId, OSLgetItem(profileReq, "profile"))
		OSLsetItem(userIdToUsername, userId, username)
		OSLsetItem(usernameToUserId, username, userId)
		fs.MkdirAll(("db/" + OSLtoString(userId)))
		fs.WriteFile((("db/" + OSLtoString(userId)) + "/user.json"), OSLtoString(OSLgetItem(profileReq, "profile")))
		return profileReq
}
func getUserIdFromUsername(username string) string{
		if OSLcontains(usernameToUserId, username) {
				return OSLtoString(OSLgetItem(usernameToUserId, username))
		}
		return ""
}
func userDbPath(userId string) string{
		return (("db/" + OSLtoString(userId)) + "/images.json")
}
func ensureUserDb(userId string) {
	var dir string
	var path string
		dir = ("db/" + OSLtoString(userId))
		if (fs.Exists(dir) != true) {
				fs.MkdirAll(dir)
		}
		path = userDbPath(userId)
		if (fs.Exists(path) != true) {
				fs.WriteFile(path, "[]")
		}
		ensureCacheDir(userId)
}
func readUserImages(userId string) []any{
	var path string
	var data string
	var arr []any
		ensureUserDb(userId)
		path = userDbPath(userId)
		data = OSLtoString(fs.ReadFile(path))
		if OSLcastBool(OSLequal(data, "")) {
				return []any{}
		}
		arr = OSLcastArray(JsonParse(data))
		return arr
}
func enrichImagesWithSharing(userId string, images []any) []any{
	var sharesObj map[string]any
	var shares []any
	var img map[string]any
	var share map[string]any
		sharesObj = readUserShares(userId)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		for i := 1; i <= OSLlen(images); i++ {
				img = OSLcastObject(OSLgetItem(images, i))
				for j := 1; j <= OSLlen(shares); j++ {
						share = OSLcastObject(OSLgetItem(shares, j))
						if OSLcastBool(OSLequal(OSLgetItem(share, "imageId"), OSLgetItem(img, "id"))) {
								OSLsetItem(img, "sharedWith", OSLgetItem(share, "sharedWith"))
								OSLsetItem(img, "isPublic", OSLgetItem(share, "isPublic"))
						}
				}
		}
		return images
}
func writeUserImages(userId string, arr []any) bool{
	var path string
	var out string
		path = userDbPath(userId)
		out = OSLtoString(arr)
		return fs.WriteFile(path, out)
}
func findImage(arr []any, id string) map[string]any{
	var it map[string]any
		for i := 1; i <= OSLlen(arr); i++ {
				it = OSLcastObject(OSLgetItem(arr, i))
				if OSLcastBool(OSLequal(OSLgetItem(it, "id"), id)) {
						return it
				}
		}
		return map[string]any{}
}
func removeImage(arr []any, id string) []any{
	var out []any
	var it map[string]any
		out = []any{}
		for i := 1; i <= OSLlen(arr); i++ {
				it = OSLcastObject(OSLgetItem(arr, i))
				if OSLcastBool(OSLnotEqual(OSLgetItem(it, "id"), id)) {
						OSLappend(&(out), it)
				}
		}
		return out
}
func removeImages(arr []any, ids []any) []any{
	var out []any
	var it map[string]any
		out = []any{}
		for i := 1; i <= OSLlen(arr); i++ {
				it = OSLcastObject(OSLgetItem(arr, i))
				if (OSLcontains(ids, OSLgetItem(it, "id")) != true) {
						OSLappend(&(out), it)
				}
		}
		return out
}
func userAlbumsPath(userId string) string{
		return (("db/" + OSLtoString(userId)) + "/albums.json")
}
func ensureUserAlbums(userId string) {
	var dir string
	var path string
		dir = ("db/" + OSLtoString(userId))
		if (fs.Exists(dir) != true) {
				fs.MkdirAll(dir)
		}
		path = userAlbumsPath(userId)
		if (fs.Exists(path) != true) {
				fs.WriteFile(path, "{ \"albums\": [], \"items\": {} }")
		}
}
func readUserAlbums(userId string) map[string]any{
	var path string
	var data string
	var obj map[string]any
		ensureUserAlbums(userId)
		path = userAlbumsPath(userId)
		data = OSLtoString(fs.ReadFile(path))
		if OSLcastBool(OSLequal(data, "")) {
				return map[string]any{
						"albums": []any{},
						"items": map[string]any{},
				}
		}
		obj = OSLcastObject(JsonParse(data))
		if OSLcastBool(OSLequal(OSLgetItem(obj, "albums"), nil)) {
				OSLsetItem(obj, "albums", []any{})
		}
		if OSLcastBool(OSLequal(OSLgetItem(obj, "items"), nil)) {
				OSLsetItem(obj, "items", map[string]any{})
		}
		return obj
}
func writeUserAlbums(userId string, albums map[string]any) bool{
	var path string
	var out string
		path = userAlbumsPath(userId)
		out = OSLtoString(albums)
		return fs.WriteFile(path, out)
}
func addAlbum(userId string, name string) map[string]any{
	var albums map[string]any
	var list []any
	var exists bool
	var items map[string]any
		albums = readUserAlbums(userId)
		list = OSLcastArray(OSLgetItem(albums, "albums"))
		exists = false
		for i := 1; i <= OSLlen(list); i++ {
				if OSLcastBool(OSLequal(strings.ToLower(OSLtoString(OSLgetItem(list, i))), strings.ToLower(name))) {
						exists = true
				}
		}
		if (exists != true) {
				OSLappend(&(list), name)
				OSLsetItem(albums, "albums", list)
		}
		if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(albums, "items"), name), nil)) {
				items = OSLcastObject(OSLgetItem(albums, "items"))
				OSLsetItem(items, name, []any{})
		}
		writeUserAlbums(userId, albums)
		return albums
}
func removeAlbumDef(userId string, name string) map[string]any{
	var albums map[string]any
	var list []any
	var out []any
	var it string
		albums = readUserAlbums(userId)
		list = OSLcastArray(OSLgetItem(albums, "albums"))
		out = []any{}
		for i := 1; i <= OSLlen(list); i++ {
				it = OSLtoString(OSLgetItem(list, i))
				if OSLcastBool(OSLnotEqual(strings.ToLower(it), strings.ToLower(name))) {
						OSLappend(&(out), it)
				}
		}
		OSLsetItem(albums, "albums", out)
		OSLdelete(OSLgetItem(albums, "items"), name)
		writeUserAlbums(userId, albums)
		return albums
}
func addImageToAlbum(userId string, name string, id string) map[string]any{
	var albums map[string]any
	var items map[string]any
	var ids []any
	var exists bool
		albums = readUserAlbums(userId)
		items = OSLcastObject(OSLgetItem(albums, "items"))
		if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(albums, "items"), name), nil)) {
				OSLsetItem(items, name, []any{})
		}
		ids = OSLcastArray(OSLgetItem(items, name))
		exists = false
		for i := 1; i <= OSLlen(ids); i++ {
				if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(ids, i)), id)) {
						exists = true
				}
		}
		if (exists != true) {
				OSLappend(&(ids), id)
				OSLsetItem(items, name, ids)
				writeUserAlbums(userId, albums)
		}
		return albums
}
func removeImageFromAlbum(userId string, name string, id string) map[string]any{
	var albums map[string]any
	var ids []any
	var out []any
	var items map[string]any
		albums = readUserAlbums(userId)
		ids = OSLcastArray(OSLgetItem(OSLgetItem(albums, "items"), name))
		out = []any{}
		for i := 1; i <= OSLlen(ids); i++ {
				if OSLcastBool(OSLnotEqual(OSLtoString(OSLgetItem(ids, i)), id)) {
						OSLappend(&(out), OSLgetItem(ids, i))
				}
		}
		items = OSLcastObject(OSLgetItem(albums, "items"))
		OSLsetItem(items, name, out)
		writeUserAlbums(userId, albums)
		return albums
}
func userBinPath(userId string) string{
		return (("db/" + OSLtoString(userId)) + "/bin.json")
}
func ensureUserBin(userId string) {
	var dir string
	var path string
		dir = ("db/" + OSLtoString(userId))
		if (fs.Exists(dir) != true) {
				fs.MkdirAll(dir)
		}
		path = userBinPath(userId)
		if (fs.Exists(path) != true) {
				fs.WriteFile(path, "[]")
		}
}
func ensureCacheDir(userId string) {
	var dir string
		dir = (("db/" + OSLtoString(userId)) + "/cache")
		if (fs.Exists(dir) != true) {
				fs.MkdirAll(dir)
		}
}
func readUserBin(userId string) []any{
	var path string
	var data string
	var arr []any
		ensureUserBin(userId)
		path = userBinPath(userId)
		data = OSLtoString(fs.ReadFile(path))
		if OSLcastBool(OSLequal(data, "")) {
				return []any{}
		}
		arr = OSLcastArray(JsonParse(data))
		return arr
}
func writeUserBin(userId string, arr []any) bool{
	var path string
	var out string
		path = userBinPath(userId)
		out = OSLtoString(arr)
		return fs.WriteFile(path, out)
}
func calculateStorageStats(userId string) map[string]any{
	var path string
	var files []any
	var totalBytes float64
	var imageCount int
	var entries []any
	var sizeGroups map[string]any
	var name string
	var fpath string
	var stat map[string]any
	var size float64
	var id string
	var sizeStr string
	var sg []any
	var hashGroups map[string]any
	var duplicateGroups []any
	var sizeKeys []any
	var sKey string
	var group []any
	var it map[string]any
	var h string
	var hg []any
	var hashKeys []any
	var hKey string
	var n int
	var tmp map[string]any
	var largest []any
	var limit int
	var binPath string
	var binBytes float64
	var binFiles []any
	var bname string
	var bfpath string
	var bstat map[string]any
	var fileSizes map[string]any
		path = (("db/" + OSLtoString(userId)) + "/blob")
		if (fs.Exists(path) != true) {
				return map[string]any{
						"totalBytes": 0,
						"imageCount": 0,
						"largestImages": []any{},
						"binBytes": 0,
						"duplicateGroups": []any{},
				}
		}
		files = OSLcastArray(fs.ReadDir(path))
		totalBytes = 0
		imageCount = int(0)
		entries = []any{}
		sizeGroups = map[string]any{}
		for i := 1; i <= OSLlen(files); i++ {
				name = OSLtoString(OSLgetItem(files, i))
				if strings.HasSuffix(name, ".jpg") {
						fpath = ((OSLtoString(path) + "/") + OSLtoString(name))
						stat = OSLcastObject(fs.GetStat(fpath))
						size = OSLcastNumber(OSLgetItem(stat, "size"))
						totalBytes = OSLadd(totalBytes, size)
						imageCount = OSLadd(imageCount, 1)
						id = OSLtoString(OSLtrim(name, 1, -5))
						sizeStr = OSLtoString(size)
						if OSLcastBool(OSLequal(OSLgetItem(sizeGroups, sizeStr), nil)) {
								OSLsetItem(sizeGroups, sizeStr, []any{})
						}
						sg = OSLcastArray(OSLgetItem(sizeGroups, sizeStr))
						OSLappend(&(sg), map[string]any{
								"id": id,
								"path": fpath,
						})
						OSLsetItem(sizeGroups, sizeStr, sg)
						OSLappend(&(entries), map[string]any{
								"id": id,
								"bytes": size,
								"path": fpath,
						})
				}
		}
		hashGroups = map[string]any{}
		duplicateGroups = []any{}
		sizeKeys = OSLgetKeys(sizeGroups)
		for i := 1; i <= OSLlen(sizeKeys); i++ {
				sKey = OSLtoString(OSLgetItem(sizeKeys, i))
				group = OSLcastArray(OSLgetItem(sizeGroups, sKey))
				if OSLcastBool(OSLcastNumber(OSLlen(group)) > OSLcastNumber(1)) {
						for j := 1; j <= OSLlen(group); j++ {
								it = OSLcastObject(OSLgetItem(group, j))
								h = calculateFileHash(OSLtoString(OSLgetItem(it, "path")))
								if OSLcastBool(OSLequal(OSLgetItem(hashGroups, h), nil)) {
										OSLsetItem(hashGroups, h, []any{})
								}
								hg = OSLcastArray(OSLgetItem(hashGroups, h))
								OSLappend(&(hg), OSLtoString(OSLgetItem(it, "id")))
								OSLsetItem(hashGroups, h, hg)
						}
				}
		}
		hashKeys = OSLgetKeys(hashGroups)
		for i := 1; i <= OSLlen(hashKeys); i++ {
				hKey = OSLtoString(OSLgetItem(hashKeys, i))
				group = OSLcastArray(OSLgetItem(hashGroups, hKey))
				if OSLcastBool(OSLcastNumber(OSLlen(group)) > OSLcastNumber(1)) {
						OSLappend(&(duplicateGroups), map[string]any{
								"hash": hKey,
								"ids": group,
						})
				}
		}
		n = OSLlen(entries)
		for i := 1; i <= OSLround(n); i++ {
				for j := 1; j <= int((OSLsub(n, i) - 1)); j++ {
						if OSLcastBool(OSLcastNumber(OSLcastNumber(OSLgetItem(OSLgetItem(entries, j), "bytes"))) < OSLcastNumber(OSLcastNumber(OSLgetItem(OSLgetItem(entries, OSLadd(j, 1)), "bytes")))) {
								tmp = OSLcastObject(OSLgetItem(entries, j))
								OSLsetItem(entries, j, OSLgetItem(entries, OSLadd(j, 1)))
								OSLsetItem(entries, OSLadd(j, 1), tmp)
						}
				}
		}
		largest = []any{}
		limit = int(10)
		if OSLcastBool(OSLcastNumber(n) < OSLcastNumber(limit)) {
				limit = n
		}
		for i := 1; i <= OSLround(limit); i++ {
				OSLappend(&(largest), map[string]any{
						"id": OSLgetItem(OSLgetItem(entries, i), "id"),
						"bytes": OSLgetItem(OSLgetItem(entries, i), "bytes"),
				})
		}
		binPath = (("db/" + OSLtoString(userId)) + "/bin")
		binBytes = 0
		if OSLcastBool(fs.Exists(binPath)) {
				binFiles = OSLcastArray(fs.ReadDir(binPath))
				for i := 1; i <= OSLlen(binFiles); i++ {
						bname = OSLtoString(OSLgetItem(binFiles, i))
						if strings.HasSuffix(bname, ".jpg") {
								bfpath = ((OSLtoString(binPath) + "/") + OSLtoString(bname))
								bstat = OSLcastObject(fs.GetStat(bfpath))
								binBytes = OSLadd(binBytes, OSLcastNumber(OSLgetItem(bstat, "size")))
						}
				}
		}
		fileSizes = map[string]any{}
		for i := 1; i <= OSLlen(entries); i++ {
				OSLsetItem(fileSizes, OSLtoString(OSLgetItem(OSLgetItem(entries, i), "id")), OSLgetItem(OSLgetItem(entries, i), "bytes"))
		}
		return map[string]any{
				"totalBytes": totalBytes,
				"imageCount": imageCount,
				"largestImages": largest,
				"binBytes": binBytes,
				"duplicateGroups": duplicateGroups,
				"fileSizes": fileSizes,
		}
}
func userSharesPath(userId string) string{
		return (("db/" + OSLtoString(userId)) + "/shares.json")
}
func ensureUserShares(userId string) {
	var dir string
	var path string
		dir = ("db/" + OSLtoString(userId))
		if (fs.Exists(dir) != true) {
				fs.MkdirAll(dir)
		}
		path = userSharesPath(userId)
		if (fs.Exists(path) != true) {
				fs.WriteFile(path, "{ \"shares\": [] }")
		}
}
func readUserShares(userId string) map[string]any{
	var path string
	var data string
	var obj map[string]any
		ensureUserShares(userId)
		path = userSharesPath(userId)
		data = OSLtoString(fs.ReadFile(path))
		if OSLcastBool(OSLequal(data, "")) {
				return map[string]any{
						"shares": []any{},
				}
		}
		obj = OSLcastObject(JsonParse(data))
		if OSLcastBool(OSLequal(OSLgetItem(obj, "shares"), nil)) {
				OSLsetItem(obj, "shares", []any{})
		}
		return obj
}
func writeUserShares(userId string, sharesObj map[string]any) bool{
	var path string
		path = userSharesPath(userId)
		return fs.WriteFile(path, OSLtoString(sharesObj))
}
func addShare(owner string, imageId string, targetUserId string) map[string]any{
	var sharesObj map[string]any
	var shares []any
	var found bool
	var share map[string]any
	var sharedWith []any
	var exists bool
		sharesObj = readUserShares(owner)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		found = false
		for i := 1; i <= OSLlen(shares); i++ {
				share = OSLcastObject(OSLgetItem(shares, i))
				if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(share, "imageId")), imageId)) {
						sharedWith = OSLcastArray(OSLgetItem(share, "sharedWith"))
						exists = false
						for j := 1; j <= OSLlen(sharedWith); j++ {
								if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(sharedWith, j)), targetUserId)) {
										exists = true
								}
						}
						if (exists != true) {
								OSLappend(&(sharedWith), targetUserId)
								OSLsetItem(share, "sharedWith", sharedWith)
						}
						found = true
				}
		}
		if (found != true) {
				OSLappend(&(shares), map[string]any{
						"imageId": imageId,
						"sharedWith": []any{
								targetUserId,
						},
				})
		}
		OSLsetItem(sharesObj, "shares", shares)
		writeUserShares(owner, sharesObj)
		return sharesObj
}
func removeShare(owner string, imageId string, targetUserId string) map[string]any{
	var sharesObj map[string]any
	var shares []any
	var newShares []any
	var share map[string]any
	var sharedWith []any
	var newSharedWith []any
		sharesObj = readUserShares(owner)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		newShares = []any{}
		for i := 1; i <= OSLlen(shares); i++ {
				share = OSLcastObject(OSLgetItem(shares, i))
				if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(share, "imageId")), imageId)) {
						sharedWith = OSLcastArray(OSLgetItem(share, "sharedWith"))
						newSharedWith = []any{}
						for j := 1; j <= OSLlen(sharedWith); j++ {
								if OSLcastBool(OSLnotEqual(OSLtoString(OSLgetItem(sharedWith, j)), targetUserId)) {
										OSLappend(&(newSharedWith), OSLgetItem(sharedWith, j))
								}
						}
						if OSLcastBool(OSLcastNumber(OSLlen(newSharedWith)) > OSLcastNumber(0)) {
								OSLsetItem(share, "sharedWith", newSharedWith)
								OSLappend(&(newShares), share)
						}
				} else {
						OSLappend(&(newShares), share)
				}
		}
		OSLsetItem(sharesObj, "shares", newShares)
		writeUserShares(owner, sharesObj)
		return sharesObj
}
func setPublicShare(owner string, imageId string, isPublic bool) map[string]any{
	var sharesObj map[string]any
	var shares []any
	var found bool
	var share map[string]any
		sharesObj = readUserShares(owner)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		found = false
		for i := 1; i <= OSLlen(shares); i++ {
				share = OSLcastObject(OSLgetItem(shares, i))
				if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(share, "imageId")), imageId)) {
						OSLsetItem(share, "isPublic", isPublic)
						found = true
				}
		}
		if (found != true) && OSLequal(isPublic, true) {
				OSLappend(&(shares), map[string]any{
						"imageId": imageId,
						"sharedWith": []any{},
						"isPublic": true,
				})
		}
		OSLsetItem(sharesObj, "shares", shares)
		writeUserShares(owner, sharesObj)
		return sharesObj
}
func getSharedWithMe(userId string) []any{
	var results []any
	var dbPath string
	var dirs []any
	var ownerDir string
	var sharesPath string
	var data string
	var sharesObj map[string]any
	var shares []any
	var share map[string]any
	var sharedWith []any
		results = []any{}
		dbPath = "db"
		if (fs.Exists(dbPath) != true) {
				return results
		}
		dirs = OSLcastArray(fs.ReadDir(dbPath))
		for i := 1; i <= OSLlen(dirs); i++ {
				ownerDir = OSLtoString(OSLgetItem(dirs, i))
				sharesPath = (((OSLtoString(dbPath) + "/") + OSLtoString(ownerDir)) + "/shares.json")
				if OSLcastBool(fs.Exists(sharesPath)) {
						data = OSLtoString(fs.ReadFile(sharesPath))
						if OSLcastBool(OSLnotEqual(data, "")) {
								sharesObj = OSLcastObject(JsonParse(data))
								shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
								if OSLcastBool(OSLnotEqual(shares, nil)) {
										for j := 1; j <= OSLlen(shares); j++ {
												share = OSLcastObject(OSLgetItem(shares, j))
												sharedWith = OSLcastArray(OSLgetItem(share, "sharedWith"))
												for k := 1; k <= OSLlen(sharedWith); k++ {
														if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(sharedWith, k)), userId)) {
																OSLappend(&(results), map[string]any{
																		"owner": ownerDir,
																		"imageId": OSLgetItem(share, "imageId"),
																})
														}
												}
										}
								}
						}
				}
		}
		return results
}
func getCachePath(userId string, id string) string{
		return (((("db/" + OSLtoString(userId)) + "/cache/") + OSLtoString(id)) + "_preview.jpg")
}
func servePreview(c any, path string, id string, username string) bool{
	var cachePath string
	var cacheInfo map[string]any
	var origInfo map[string]any
	var etag string
	var ifNoneMatch string
	var fileInfo map[string]any
	var fileSize float64
	var w float64
	var h float64
	var maxw float64
	var maxh float64
	var ratio float64
	var rw int
	var rh int
	var cached bool
	var out []byte
		if (fs.Exists(path) != true) || fs.IsDir(path) {
				if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
						c.Status(404)
						return false
				}
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return false
		}
		if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
				c.Status(200)
				return true
		}
		cachePath = getCachePath(username, id)
		if OSLcastBool(fs.Exists(cachePath)) {
				cacheInfo = OSLcastObject(fs.GetStat(cachePath))
				origInfo = OSLcastObject(fs.GetStat(path))
				if OSLcastBool(OSLcastNumber(OSLcastNumber(OSLgetItem(cacheInfo, "modTime"))) >= OSLcastNumber(OSLcastNumber(OSLgetItem(origInfo, "modTime")))) {
						etag = ("preview-" + OSLtoString(OSLgetItem(cacheInfo, "modTime")))
						ifNoneMatch = OSLtoString(c.GetHeader("If-None-Match"))
						if OSLcastBool(OSLequal(ifNoneMatch, etag)) {
								c.Status(304)
								return true
						}
						c.Header("ETag", etag)
						c.Header("Cache-Control", "private, max-age=31536000, immutable")
						c.File(cachePath)
						return true
				}
		}
		fileInfo = OSLcastObject(fs.GetStat(path))
		fileSize = OSLcastNumber(OSLgetItem(fileInfo, "size"))
		if OSLcastBool(OSLequal(fileSize, 0)) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return false
		}
		if OSLcastBool(OSLcastNumber(fileSize) > OSLcastNumber(1e+07)) {
				c.JSON(413, map[string]any{
						"ok": false,
						"error": "file too large",
				})
				return false
		}
		var im = img.Open(path)
		if OSLcastBool(OSLequal(im, nil)) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "invalid image",
				})
				return false
		}
		defer im.Close()
		w = OSLcastNumber(im.Width())
		h = OSLcastNumber(im.Height())
		maxw = 200
		maxh = 200
		if OSLcastNumber(w) <= OSLcastNumber(maxw) && OSLcastNumber(h) <= OSLcastNumber(maxh) {
				c.Header("Cache-Control", "private, max-age=31536000, immutable")
				c.File(path)
				return true
		}
		ratio = OSLmin(OSLdivide(maxw, w), OSLdivide(maxh, h))
		rw = OSLround(OSLmultiply(w, ratio))
		rh = OSLround(OSLmultiply(h, ratio))
		var resized = img.ResizeFast(im, rw, rh)
		if OSLcastBool(OSLequal(resized, nil)) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "resize failed",
				})
				return false
		}
		defer resized.Close()
		cached = bool(img.SaveJPEG(resized, cachePath, 80))
		if OSLcastBool(cached) {
				c.Header("Cache-Control", "private, max-age=31536000, immutable")
				c.File(cachePath)
		} else {
				out = img.EncodeJPEGBytes(resized, 80)
				if OSLcastBool(OSLequal(OSLlen(out), 0)) {
						c.JSON(500, map[string]any{
								"ok": false,
								"error": "encode failed",
						})
						return false
				}
				c.Header("Cache-Control", "private, max-age=3600")
				c.Data(200, "image/jpeg", out)
		}
		return true
}

func handleAuth(c any) {
	var token string
	var resp map[string]any
	var json map[string]any
	var sessionId string
	var userId string
	var username string
	var profileReq map[string]any
		token = OSLtoString(c.Query("v"))
		if OSLcastBool(OSLequal(token, "")) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "missing token",
				})
				return
		}
		resp = OSLcastObject(requests.Get(("https://api.rotur.dev/validate?key=rotur-photos&v=" + OSLtoString(token))))
		if (OSLgetItem(resp, "success") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "failed to validate token",
				})
				return
		}
		json = OSLcastObject(JsonParse(OSLtoString(OSLgetItem(resp, "body"))))
		if OSLcastBool(OSLnotEqual(OSLgetItem(json, "error"), nil)) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": OSLgetItem(json, "error"),
				})
				return
		}
		if (OSLgetItem(json, "valid") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "invalidate token",
				})
				return
		}
		sessionId = randomString(32)
		userId = OSLtoString(OSLgetItem(json, "id"))
		username = OSLtoString(OSLgetItem(json, "username"))
		OSLsetItem(sessions, sessionId, userId)
		profileReq = writeProfile(userId, username)
		if (OSLgetItem(profileReq, "ok") != true) {
				c.JSON(401, OSLgetItem(profileReq, "error"))
				return
		}
		fs.WriteFile("db/sessions.json", JsonFormat(sessions))
		c.SetCookie("session_id", sessionId, 86400, "/", "", false, true)
		c.JSON(200, map[string]any{
				"ok": true,
				"sessionId": sessionId,
				"token": token,
		})
}
func handleLogout(c any) {
	var sessionId string
		sessionId = OSLtoString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
		if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
				OSLdelete(sessions, sessionId)
				fs.WriteFile("db/sessions.json", JsonFormat(sessions))
		}
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleAble(c any) {
	var userId string
	var able map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		c.JSON(200, able)
}
func handleStorage(c any) {
	var userId string
	var able map[string]any
	var stats map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		stats = calculateStorageStats(userId)
		OSLsetItem(stats, "quota", OSLgetItem(able, "storageQuota"))
		c.JSON(200, stats)
}
func handleAllImages(c any) {
	var userId string
	var able map[string]any
	var images []any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		images = readUserImages(userId)
		images = enrichImagesWithSharing(userId, images)
		c.JSON(200, images)
}
func handleRecentImages(c any) {
	var userId string
	var able map[string]any
	var arr []any
	var n int
	var nowMs float64
	var ninety float64
	var out []any
	var it map[string]any
	var ts float64
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		arr = readUserImages(userId)
		n = OSLlen(arr)
		nowMs = OSLcastNumber(time.Now().UnixMilli())
		ninety = OSLcastNumber(OSLmultiply(OSLmultiply(OSLmultiply(OSLmultiply(90, 24), 60), 60), 1000))
		out = []any{}
		for i := 1; i <= OSLround(n); i++ {
				it = OSLcastObject(OSLgetItem(arr, i))
				ts = OSLcastNumber(OSLgetItem(it, "timestamp"))
				if OSLcastBool(OSLcastNumber(OSLsub(nowMs, ts)) <= OSLcastNumber(ninety)) {
						OSLappend(&(out), it)
				}
		}
		out = enrichImagesWithSharing(userId, out)
		c.JSON(200, out)
}
func handleYearImages(c any) {
	var userId string
	var able map[string]any
	var year int
	var arr []any
	var n int
	var out []any
	var it map[string]any
	var ts int
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		year = OSLcastInt(c.Param("year"))
		arr = readUserImages(userId)
		n = OSLlen(arr)
		out = []any{}
		for i := 1; i <= OSLround(n); i++ {
				it = OSLcastObject(OSLgetItem(arr, i))
				ts = OSLcastInt(OSLgetItem(it, "timestamp"))
				var t = time.UnixMilli(int64(ts))
				if OSLcastBool(OSLequal(t.Year(), year)) {
						OSLappend(&(out), it)
				}
		}
		out = enrichImagesWithSharing(userId, out)
		c.JSON(200, out)
}
func handleMonthImages(c any) {
	var userId string
	var able map[string]any
	var year int
	var month int
	var arr []any
	var n int
	var out []any
	var it map[string]any
	var ts int
	var tMonth int
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		year = OSLcastInt(c.Param("year"))
		month = OSLcastInt(c.Param("month"))
		arr = readUserImages(userId)
		n = OSLlen(arr)
		out = []any{}
		for i := 1; i <= OSLround(n); i++ {
				it = OSLcastObject(OSLgetItem(arr, i))
				ts = OSLcastInt(OSLgetItem(it, "timestamp"))
				var t = time.UnixMilli(int64(ts))
				tMonth = int(0)
				tMonth = int(t.Month())
				if OSLequal(t.Year(), year) && OSLequal(tMonth, month) {
						OSLappend(&(out), it)
				}
		}
		out = enrichImagesWithSharing(userId, out)
		c.JSON(200, out)
}
func handleUpload(c any) {
	var userId string
	var able map[string]any
	var base string
	var tmpPath string
	var body []byte
	var stats map[string]any
	var size map[string]any
	var w int
	var h int
	var ratio float64
	var im any
	var id string
	var outPath string
	var tsms float64
	var x *exif.Exif
	var entry map[string]any
	var images []any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		base = (("db/" + OSLtoString(userId)) + "/blob")
		if (fs.Exists(base) != true) {
				fs.MkdirAll(base)
		}
		tmpPath = ((OSLtoString(base) + "/tmp-") + randomString(16))
		body = OSLgetItem(OSLcastArray(c.GetRawData()), 1).([]byte)
		if OSLcastBool(OSLequal(OSLlen(body), 0)) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "empty body",
				})
				return
		}
		stats = calculateStorageStats(userId)
		if OSLcastBool(OSLcastNumber(((OSLcastNumber(OSLgetItem(stats, "totalBytes")) + OSLcastNumber(OSLgetItem(stats, "binBytes"))) + float64(OSLlen(body)))) > OSLcastNumber(OSLcastNumber(OSLgetItem(able, "storageQuota")))) {
				c.JSON(403, map[string]any{
						"ok": false,
						"error": "storage quota exceeded",
				})
				return
		}
		if (fs.WriteFile(tmpPath, body) != true) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to write upload",
				})
				return
		}
		var im = img.Open(tmpPath)
		if OSLcastBool(OSLequal(im, nil)) {
				fs.Remove(tmpPath)
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "invalid image",
				})
				return
		}
		size = OSLcastObject(im.Size())
		w = OSLround(OSLgetItem(size, "w"))
		h = OSLround(OSLgetItem(size, "h"))
		if OSLcastNumber(w) > OSLcastNumber(OSLgetItem(downscaleWhen, "width")) || OSLcastNumber(h) > OSLcastNumber(OSLgetItem(downscaleWhen, "height")) {
				ratio = OSLcastNumber(OSLmin(OSLdivide(OSLcastInt(OSLgetItem(downscaleWhen, "width")), w), OSLdivide(OSLcastInt(OSLgetItem(downscaleWhen, "height")), h)))
				w = OSLround(OSLmultiply(w, ratio))
				h = OSLround(OSLmultiply(h, ratio))
				var resized = img.Resize(im, w, h)
				im.Close()
				if OSLcastBool(OSLequal(resized, nil)) {
						fs.Remove(tmpPath)
						c.JSON(500, map[string]any{
								"ok": false,
								"error": "resize failed",
						})
						return
				}
				im = resized
		}
		id = randomString(24)
		outPath = (((OSLtoString(base) + "/") + OSLtoString(id)) + ".jpg")
		if (img.SaveJPEG(im, outPath, 90) != true) {
				im.Close()
				fs.Remove(tmpPath)
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "save failed",
				})
				return
		}
		im.Close()
		tsms = OSLcastNumber(time.Now().UnixMilli())
		x = nil
		f, err := os.Open(tmpPath)
		if OSLcastBool(OSLequal(err, nil)) {
				defer f.Close()
				x, err = exif.Decode(f)
				if OSLcastBool(OSLequal(err, nil)) {
						t, err := x.DateTime()
						if OSLequal(err, nil) && (t.IsZero() != true) {
								tsms = OSLcastNumber(t.UnixMilli())
						}
				}
		}
		entry = map[string]any{
				"id": id,
				"width": w,
				"height": h,
				"timestamp": tsms,
				"make": "",
				"model": "",
				"exposure_time": "",
				"f_number": "",
		}
		if OSLcastBool(OSLnotEqual(x, nil)) {
				if tag, err := x.Get(exif.Make); err == nil { entry["make"], _ = tag.StringVal() }
				if tag, err := x.Get(exif.Model); err == nil { entry["model"], _ = tag.StringVal() }
				if tag, err := x.Get(exif.ExposureTime); err == nil { entry["exposure_time"], _ = tag.StringVal() }
				if tag, err := x.Get(exif.FNumber); err == nil { entry["f_number"], _ = tag.StringVal() }
		}
		fs.Remove(tmpPath)
		images = readUserImages(userId)
		OSLappend(&(images), entry)
		writeUserImages(userId, images)
		if OSLcastBool(OSLequal(c.Query("public"), "true")) {
				setPublicShare(userId, id, true)
		}
		c.JSON(200, map[string]any{
				"ok": true,
				"id": id,
				"path": ((OSLtoString(userId) + "/") + OSLtoString(id)),
				"owner": map[string]any{
						"id": userId,
						"username": OSLtoString(c.MustGet("username")),
				},
		})
}
func handleId(c any) {
	var userId string
	var able map[string]any
	var id string
	var json bool
	var it map[string]any
	var path string
	var fileInfo map[string]any
	var etag string
	var ifNoneMatch string
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		c.Header("Content-Type", "image/jpeg")
		c.Header("ETag", etag)
		c.Header("Cache-Control", "private, must-revalidate")
		id = OSLtoString(c.Param("id"))
		json = false
		if strings.HasSuffix(id, ".json") {
				id = OSLtrim(id, 1, -6)
				json = true
		}
		if OSLcastBool(json) {
				it = findImage(readUserImages(userId), id)
				if OSLcastBool(OSLequal(OSLgetItem(it, "id"), nil)) {
						c.JSON(404, map[string]any{
								"ok": false,
								"error": "not found",
						})
						return
				}
				c.JSON(200, it)
				return
		}
		path = (((("db/" + OSLtoString(userId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		if (fs.Exists(path) != true) || fs.IsDir(path) {
				if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
						c.Status(404)
						return
				}
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
				c.Status(200)
				return
		}
		fileInfo = OSLcastObject(fs.GetStat(path))
		etag = (("\"" + OSLtoString(OSLgetItem(fileInfo, "modTime"))) + "\"")
		ifNoneMatch = OSLtoString(c.GetHeader("If-None-Match"))
		if OSLcastBool(OSLequal(ifNoneMatch, etag)) {
				c.Status(304)
				return
		}
		c.File(path)
}
func handleImageRotate(c any) {
	var userId string
	var able map[string]any
	var id string
	var body map[string]any
	var angle float64
	var path string
	var wrote bool
	var arr []any
	var it map[string]any
	var oldW float64
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		body = OSLcastObject(JsonParse(OSLtoString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
		angle = OSLcastNumber(OSLgetItem(body, "angle"))
		path = (((("db/" + OSLtoString(userId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		if (fs.Exists(path) != true) || fs.IsDir(path) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		var imgFile = img.Open(path)
		if OSLcastBool(OSLequal(imgFile, nil)) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "invalid image",
				})
				return
		}
		defer imgFile.Close()
		var rotated = img.Rotate(imgFile, angle)
		if OSLcastBool(OSLequal(rotated, nil)) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to rotate",
				})
				return
		}
		defer rotated.Close()
		wrote = bool(img.SaveJPEG(rotated, path, 90))
		if (wrote != true) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to save",
				})
				return
		}
		arr = readUserImages(userId)
		it = findImage(arr, id)
		if OSLcastBool(OSLnotEqual(OSLgetItem(it, "id"), nil)) {
				if OSLequal(angle, 90) || OSLequal(angle, 270) {
						oldW = OSLcastNumber(OSLgetItem(it, "width"))
						OSLsetItem(it, "width", OSLgetItem(it, "height"))
						OSLsetItem(it, "height", oldW)
						writeUserImages(userId, arr)
				}
		}
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleImageCompress(c any) {
	var userId string
	var able map[string]any
	var id string
	var body map[string]any
	var quality float64
	var path string
	var size map[string]any
	var wrote bool
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		body = OSLcastObject(JsonParse(OSLtoString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
		quality = OSLcastNumber(OSLround(OSLcastNumber(OSLgetItem(body, "quality"))))
		if OSLcastNumber(quality) < OSLcastNumber(10) || OSLcastNumber(quality) > OSLcastNumber(100) {
				quality = 85
		}
		path = (((("db/" + OSLtoString(userId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		if (fs.Exists(path) != true) || fs.IsDir(path) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		var imgFile = img.Open(path)
		if OSLcastBool(OSLequal(imgFile, nil)) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "invalid image",
				})
				return
		}
		defer imgFile.Close()
		size = OSLcastObject(imgFile.Size())
		var w = OSLround(OSLgetItem(size, "w"))
		var h = OSLround(OSLgetItem(size, "h"))
		var resized = img.Resize(imgFile, w, h)
		defer resized.Close()
		wrote = bool(img.SaveJPEG(resized, path, OSLround(quality)))
		if (wrote != true) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to save",
				})
				return
		}
		c.JSON(200, map[string]any{
				"ok": true,
				"quality": quality,
		})
}
func handleImageResize(c any) {
	var userId string
	var able map[string]any
	var id string
	var body map[string]any
	var w int
	var h int
	var path string
	var wrote bool
	var arr []any
	var it map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		body = OSLcastObject(JsonParse(OSLtoString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
		w = OSLround(OSLgetItem(body, "width"))
		h = OSLround(OSLgetItem(body, "height"))
		if OSLcastNumber(w) <= OSLcastNumber(0) || OSLcastNumber(h) <= OSLcastNumber(0) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "invalid dimensions",
				})
				return
		}
		path = (((("db/" + OSLtoString(userId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		if (fs.Exists(path) != true) || fs.IsDir(path) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		var imgFile = img.Open(path)
		if OSLcastBool(OSLequal(imgFile, nil)) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "invalid image",
				})
				return
		}
		defer imgFile.Close()
		var resized = img.Resize(imgFile, w, h)
		if OSLcastBool(OSLequal(resized, nil)) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to resize",
				})
				return
		}
		defer resized.Close()
		wrote = bool(img.SaveJPEG(resized, path, 90))
		if (wrote != true) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to save",
				})
				return
		}
		arr = readUserImages(userId)
		it = findImage(arr, id)
		if OSLcastBool(OSLnotEqual(OSLgetItem(it, "id"), nil)) {
				OSLsetItem(it, "width", w)
				OSLsetItem(it, "height", h)
				writeUserImages(userId, arr)
		}
		c.JSON(200, map[string]any{
				"ok": true,
				"width": w,
				"height": h,
		})
}
func handleImageInfo(c any) {
	var userId string
	var able map[string]any
	var id string
	var it map[string]any
	var path string
	var stat map[string]any
	var fileSize float64
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		it = findImage(readUserImages(userId), id)
		if OSLcastBool(OSLequal(OSLgetItem(it, "id"), nil)) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		path = (((("db/" + OSLtoString(userId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		if (fs.Exists(path) != true) || fs.IsDir(path) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		stat = OSLcastObject(fs.GetStat(path))
		fileSize = OSLcastNumber(OSLgetItem(stat, "size"))
		OSLsetItem(it, "fileSize", fileSize)
		c.JSON(200, it)
}
func handleSearch(c any) {
	var userId string
	var able map[string]any
	var q string
	var arr []any
	var months []any
	var albums map[string]any
	var out []any
	var n int
	var it map[string]any
	var match bool
	var ts float64
	var year int
	var month int
	var monthName string
	var make string
	var model string
	var items map[string]any
	var albumName string
	var albumIds []any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		q = strings.ToLower(c.Query("q"))
		arr = readUserImages(userId)
		if OSLcastBool(OSLequal(q, "")) {
				c.JSON(200, arr)
				return
		}
		months = []any{
				"january",
				"february",
				"march",
				"april",
				"may",
				"june",
				"july",
				"august",
				"september",
				"october",
				"november",
				"december",
		}
		albums = readUserAlbums(userId)
		out = []any{}
		n = OSLlen(arr)
		for i := 1; i <= OSLround(n); i++ {
				it = OSLcastObject(OSLgetItem(arr, i))
				match = false
				ts = OSLcastNumber(OSLgetItem(it, "timestamp"))
				if OSLcastBool(OSLcastNumber(ts) > OSLcastNumber(0)) {
						var t = time.UnixMilli(int64(ts))
						year = OSLcastInt(t.Year())
						month = OSLround(t.Month())
						if OSLcontains(q, OSLtoString(year)) {
								match = true
						}
						if OSLcastNumber(month) > OSLcastNumber(0) && OSLcastNumber(month) <= OSLcastNumber(12) {
								monthName = OSLtoString(OSLgetItem(months, month))
								if OSLcontains(q, monthName) {
										match = true
								}
						}
				}
				make = strings.ToLower(OSLtoString(OSLgetItem(it, "make")))
				model = strings.ToLower(OSLtoString(OSLgetItem(it, "model")))
				if OSLnotEqual(make, "") && OSLcontains(q, make) {
						match = true
				}
				if OSLnotEqual(model, "") && OSLcontains(q, model) {
						match = true
				}
				items = OSLcastObject(OSLgetItem(albums, "items"))
				for j := 1; j <= OSLlen(OSLgetItem(albums, "albums")); j++ {
						albumName = OSLtoString(OSLgetItem(OSLgetItem(albums, "albums"), j))
						if OSLcontains(strings.ToLower(albumName), q) {
								albumIds = OSLcastArray(OSLgetItem(items, albumName))
								for k := 1; k <= OSLlen(albumIds); k++ {
										if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(albumIds, k)), OSLtoString(OSLgetItem(it, "id")))) {
												match = true
										}
								}
						}
				}
				if OSLcastBool(match) {
						OSLappend(&(out), it)
				}
		}
		c.JSON(200, out)
}
func handleShareMine(c any) {
	var userId string
	var sharesObj map[string]any
	var shares []any
	var images []any
	var out []any
	var share map[string]any
	var imageId string
	var img map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		sharesObj = readUserShares(userId)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		images = readUserImages(userId)
		out = []any{}
		for i := 1; i <= OSLlen(shares); i++ {
				share = OSLcastObject(OSLgetItem(shares, i))
				imageId = OSLtoString(OSLgetItem(share, "imageId"))
				img = findImage(images, imageId)
				if OSLcastBool(OSLnotEqual(OSLgetItem(img, "id"), nil)) {
						OSLsetItem(img, "sharedWith", OSLgetItem(share, "sharedWith"))
						OSLappend(&(out), img)
				}
		}
		c.JSON(200, out)
}
func handleShareOthers(c any) {
	var userId string
	var sharedWithMe []any
	var results []any
	var item map[string]any
	var owner string
	var imageId string
	var ownerImages []any
	var img map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		sharedWithMe = getSharedWithMe(userId)
		results = []any{}
		for i := 1; i <= OSLlen(sharedWithMe); i++ {
				item = OSLcastObject(OSLgetItem(sharedWithMe, i))
				owner = OSLtoString(OSLgetItem(item, "owner"))
				imageId = OSLtoString(OSLgetItem(item, "imageId"))
				ownerImages = readUserImages(owner)
				img = findImage(ownerImages, imageId)
				if OSLcastBool(OSLnotEqual(OSLgetItem(img, "id"), nil)) {
						OSLsetItem(img, "owner", owner)
						OSLappend(&(results), img)
				}
		}
		c.JSON(200, results)
}
func handleSharePatch(c any) {
	var userId string
	var imageId string
	var body map[string]any
	var add []any
	var remove []any
	var targetUsername string
	var targetUserId string
	var targetOrUserId string
		userId = OSLtoString(c.MustGet("userId"))
		imageId = OSLtoString(c.Param("id"))
		body = OSLcastObject(JsonParse(OSLtoString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
		add = OSLcastArray(OSLgetItem(body, "add"))
		remove = OSLcastArray(OSLgetItem(body, "remove"))
		if OSLcastBool(OSLequal(add, nil)) {
				add = []any{}
		}
		if OSLcastBool(OSLequal(remove, nil)) {
				remove = []any{}
		}
		for i := 1; i <= OSLlen(add); i++ {
				targetUsername = OSLtoString(OSLgetItem(add, i))
				targetUserId = getUserIdFromUsername(targetUsername)
				if OSLcastBool(OSLnotEqual(targetUserId, "")) {
						addShare(userId, imageId, targetUserId)
				}
		}
		for i := 1; i <= OSLlen(remove); i++ {
				targetOrUserId = OSLtoString(OSLgetItem(remove, i))
				targetUserId = ""
				if OSLcastBool(OSLequal(OSLlen(targetOrUserId), 36)) {
						targetUserId = targetOrUserId
				} else {
						targetUserId = getUserIdFromUsername(targetOrUserId)
				}
				if OSLcastBool(OSLnotEqual(targetUserId, "")) {
						removeShare(userId, imageId, targetUserId)
				}
		}
		if OSLcastBool(OSLnotEqual(OSLgetItem(body, "isPublic"), nil)) {
				setPublicShare(userId, imageId, OSLcastBool(OSLgetItem(body, "isPublic")))
		}
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleShareCreate(c any) {
	var userId string
	var body map[string]any
	var imageId string
	var targetUser string
	var images []any
	var img map[string]any
	var targetUserId string
		userId = OSLtoString(c.MustGet("userId"))
		body = OSLcastObject(JsonParse(OSLtoString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
		imageId = OSLtoString(OSLgetItem(body, "imageId"))
		targetUser = OSLtoString(OSLgetItem(body, "username"))
		if OSLequal(imageId, "") || OSLequal(targetUser, "") {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "missing imageId or username",
				})
				return
		}
		images = readUserImages(userId)
		img = findImage(images, imageId)
		if OSLcastBool(OSLequal(OSLgetItem(img, "id"), nil)) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "image not found",
				})
				return
		}
		targetUserId = getUserIdFromUsername(targetUser)
		if OSLcastBool(OSLequal(targetUserId, "")) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "user not found",
				})
				return
		}
		addShare(userId, imageId, targetUserId)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleSharedImage(c any) {
	var ownerUsername string
	var id string
	var requestingUserId string
	var sessionId string
	var ownerId string
	var sharesObj map[string]any
	var shares []any
	var allowed bool
	var share map[string]any
	var sharedWith []any
	var path string
		ownerUsername = OSLtoString(c.Param("owner"))
		id = OSLtoString(c.Param("id"))
		requestingUserId = ""
		sessionId = OSLtoString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
		if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
				requestingUserId = OSLtoString(OSLgetItem(sessions, sessionId))
		}
		ownerId = getUserIdFromUsername(ownerUsername)
		if OSLcastBool(OSLequal(ownerId, "")) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "owner not found",
				})
				return
		}
		sharesObj = readUserShares(ownerId)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		allowed = false
		if OSLcastBool(OSLequal(requestingUserId, ownerId)) {
				allowed = true
		}
		for i := 1; i <= OSLlen(shares); i++ {
				share = OSLcastObject(OSLgetItem(shares, i))
				if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(share, "imageId")), id)) {
						if OSLcastBool(OSLequal(OSLgetItem(share, "isPublic"), true)) {
								allowed = true
								break
						}
						sharedWith = OSLcastArray(OSLgetItem(share, "sharedWith"))
						for j := 1; j <= OSLlen(sharedWith); j++ {
								if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(sharedWith, j)), requestingUserId)) {
										allowed = true
								}
						}
						if OSLcastBool(allowed) {
								break
						}
				}
		}
		if (allowed != true) {
				if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
						c.Status(404)
						return
				}
				c.JSON(403, map[string]any{
						"ok": false,
						"error": "not authorized to view this image",
				})
				return
		}
		path = (((("db/" + OSLtoString(ownerId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		if (fs.Exists(path) != true) {
				if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
						c.Status(404)
						return
				}
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
				c.Status(200)
				return
		}
		c.File(path)
}
func handleSharedImageInfo(c any) {
	var ownerUsername string
	var id string
	var requestingUserId string
	var sessionId string
	var ownerId string
	var sharesObj map[string]any
	var shares []any
	var allowed bool
	var share map[string]any
	var sharedWith []any
	var images []any
	var img map[string]any
		ownerUsername = OSLtoString(c.Param("owner"))
		id = OSLtoString(c.Param("id"))
		requestingUserId = ""
		sessionId = OSLtoString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
		if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
				requestingUserId = OSLtoString(OSLgetItem(sessions, sessionId))
		}
		ownerId = getUserIdFromUsername(ownerUsername)
		if OSLcastBool(OSLequal(ownerId, "")) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "owner not found",
				})
				return
		}
		sharesObj = readUserShares(ownerId)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		allowed = false
		if OSLcastBool(OSLequal(requestingUserId, ownerId)) {
				allowed = true
		}
		for i := 1; i <= OSLlen(shares); i++ {
				share = OSLcastObject(OSLgetItem(shares, i))
				if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(share, "imageId")), id)) {
						sharedWith = OSLcastArray(OSLgetItem(share, "sharedWith"))
						for j := 1; j <= OSLlen(sharedWith); j++ {
								if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(sharedWith, j)), requestingUserId)) {
										allowed = true
								}
						}
				}
		}
		if (allowed != true) {
				c.JSON(403, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		images = readUserImages(ownerId)
		img = findImage(images, id)
		if OSLcastBool(OSLequal(OSLgetItem(img, "id"), nil)) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		OSLsetItem(img, "owner", ownerId)
		c.JSON(200, img)
}
func handlePreview(c any) {
	var userId string
	var able map[string]any
	var id string
	var path string
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		path = (((("db/" + OSLtoString(userId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		servePreview(c, path, id, userId)
}
func handleDeleteImage(c any) {
	var userId string
	var able map[string]any
	var id string
	var base string
	var path string
	var binBase string
	var data []byte
	var s map[string]any
	var w int
	var h int
	var binPath string
	var wrote bool
	var binArr []any
	var tsms float64
	var arr []any
	var it map[string]any
	var origTs float64
	var entry map[string]any
	var out []any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		base = ("db/" + OSLtoString(userId))
		path = (((OSLtoString(base) + "/blob/") + OSLtoString(id)) + ".jpg")
		binBase = (OSLtoString(base) + "/bin")
		if (fs.Exists(binBase) != true) {
				fs.MkdirAll(binBase)
		}
		data = fs.ReadFileBytes(path)
		if OSLcastBool(OSLcastNumber(OSLlen(data)) > OSLcastNumber(0)) {
				var imgFile = img.DecodeBytes(data)
				if OSLcastBool(OSLnotEqual(imgFile, nil)) {
						s = OSLcastObject(imgFile.Size())
						w = OSLround(OSLgetItem(s, "w"))
						h = OSLround(OSLgetItem(s, "h"))
						var resized = img.Resize(imgFile, w, h)
						if OSLcastBool(OSLequal(resized, nil)) {
								c.JSON(500, map[string]any{
										"ok": false,
										"error": "resize failed",
								})
								return
						}
						binPath = (((OSLtoString(binBase) + "/") + OSLtoString(id)) + ".jpg")
						wrote = img.SaveJPEG(resized, binPath, 90)
						if OSLcastBool(wrote) {
								binArr = readUserBin(userId)
								tsms = OSLcastNumber(time.Now().UnixMilli())
								arr = readUserImages(userId)
								it = findImage(arr, id)
								origTs = OSLcastNumber(OSLgetItem(it, "timestamp"))
								if OSLcastBool(OSLequal(origTs, 0)) {
										origTs = tsms
								}
								entry = map[string]any{
										"id": id,
										"width": w,
										"height": h,
										"timestamp": origTs,
										"deletedAt": tsms,
								}
								OSLappend(&(binArr), entry)
								writeUserBin(userId, binArr)
						}
				}
		}
		if OSLcastBool(fs.Exists(path)) {
				fs.Remove(path)
		}
		arr = readUserImages(userId)
		out = removeImage(arr, id)
		writeUserImages(userId, out)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleDeleteImages(c any) {
	var userId string
	var able map[string]any
	var ids any
	var arr []any
	var out []any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		ids = OSLgetItem(JsonParse(OSLtoString(OSLgetItem(OSLgetItem(c, "request"), "body"))), "ids")
		if OSLcastBool(OSLequal(OSLlen(ids), 0)) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "missing ids",
				})
				return
		}
		if OSLcastBool(OSLnotEqual(OSLtypeof(ids), "array")) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "invalid ids",
				})
				return
		}
		for i := 1; i <= OSLlen(ids); i++ {
				fs.Remove((((("db/" + OSLtoString(userId)) + "/blob/") + OSLtoString(OSLgetItem(ids, i))) + ".jpg"))
		}
		arr = readUserImages(userId)
		out = removeImages(arr, ids.([]any))
		writeUserImages(userId, out)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleBinList(c any) {
	var userId string
	var able map[string]any
	var arr []any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		arr = readUserBin(userId)
		c.JSON(200, arr)
}
func handleBinRestore(c any) {
	var userId string
	var able map[string]any
	var id string
	var base string
	var binPath string
	var blobPath string
	var size map[string]any
	var w int
	var h int
	var wrote bool
	var binArr []any
	var newBin []any
	var ts float64
	var it map[string]any
	var arr []any
	var entry map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		base = ("db/" + OSLtoString(userId))
		binPath = (((OSLtoString(base) + "/bin/") + OSLtoString(id)) + ".jpg")
		blobPath = (((OSLtoString(base) + "/blob/") + OSLtoString(id)) + ".jpg")
		var imgFile = img.Open(binPath)
		if OSLcastBool(OSLequal(imgFile, nil)) {
				c.JSON(404, map[string]any{
						"ok": false,
						"error": "not found",
				})
				return
		}
		defer imgFile.Close()
		size = OSLcastObject(imgFile.Size())
		w = OSLround(OSLgetItem(size, "w"))
		h = OSLround(OSLgetItem(size, "h"))
		var resized = img.Resize(imgFile, w, h)
		if OSLcastBool(OSLequal(resized, nil)) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to resize",
				})
				return
		}
		defer resized.Close()
		wrote = img.SaveJPEG(resized, blobPath, 90)
		if (wrote != true) {
				c.JSON(500, map[string]any{
						"ok": false,
						"error": "failed to restore",
				})
				return
		}
		binArr = readUserBin(userId)
		newBin = []any{}
		ts = OSLcastNumber(time.Now().UnixMilli())
		for i := 1; i <= OSLlen(binArr); i++ {
				it = OSLcastObject(OSLgetItem(binArr, i))
				if OSLcastBool(OSLnotEqual(OSLtoString(OSLgetItem(it, "id")), id)) {
						OSLappend(&(newBin), it)
				} else {
						ts = OSLcastNumber(OSLgetItem(it, "timestamp"))
				}
		}
		writeUserBin(userId, newBin)
		arr = readUserImages(userId)
		entry = map[string]any{
				"id": id,
				"width": w,
				"height": h,
				"timestamp": ts,
		}
		OSLappend(&(arr), entry)
		writeUserImages(userId, arr)
		fs.Remove(binPath)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleBinDelete(c any) {
	var userId string
	var able map[string]any
	var id string
	var base string
	var binPath string
	var binArr []any
	var newBin []any
	var it map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		base = ("db/" + OSLtoString(userId))
		binPath = (((OSLtoString(base) + "/bin/") + OSLtoString(id)) + ".jpg")
		if OSLcastBool(fs.Exists(binPath)) {
				fs.Remove(binPath)
		}
		binArr = readUserBin(userId)
		newBin = []any{}
		for i := 1; i <= OSLlen(binArr); i++ {
				it = OSLcastObject(OSLgetItem(binArr, i))
				if OSLcastBool(OSLnotEqual(OSLtoString(OSLgetItem(it, "id")), id)) {
						OSLappend(&(newBin), it)
				}
		}
		writeUserBin(userId, newBin)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleBinEmpty(c any) {
	var userId string
	var able map[string]any
	var base string
	var binDir string
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		writeUserBin(userId, []any{})
		base = ("db/" + OSLtoString(userId))
		binDir = (OSLtoString(base) + "/bin")
		fs.Remove(binDir)
		fs.Mkdir(binDir)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleAlbums(c any) {
	var userId string
	var able map[string]any
	var albums map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		albums = readUserAlbums(userId)
		c.JSON(200, OSLgetItem(albums, "albums"))
}
func handleAlbumCreate(c any) {
	var userId string
	var able map[string]any
	var name string
	var albums map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		name = OSLtoString(c.Query("name"))
		if OSLcastBool(OSLequal(name, "")) {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "missing name",
				})
				return
		}
		albums = addAlbum(userId, name)
		c.JSON(200, OSLgetItem(albums, "albums"))
}
func handleAlbumDelete(c any) {
	var userId string
	var able map[string]any
	var name string
	var albums map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		name = OSLtoString(c.Param("name"))
		albums = removeAlbumDef(userId, name)
		c.JSON(200, OSLgetItem(albums, "albums"))
}
func handleAlbumImages(c any) {
	var userId string
	var able map[string]any
	var name string
	var albums map[string]any
	var ids []any
	var all []any
	var out []any
	var id string
	var it map[string]any
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		name = OSLtoString(c.Param("name"))
		albums = readUserAlbums(userId)
		ids = OSLcastArray(OSLgetItem(OSLgetItem(albums, "items"), name))
		all = readUserImages(userId)
		out = []any{}
		for i := 1; i <= OSLlen(ids); i++ {
				id = OSLtoString(OSLgetItem(ids, i))
				it = findImage(all, id)
				if OSLcastBool(OSLnotEqual(OSLgetItem(it, "id"), nil)) {
						OSLappend(&(out), it)
				}
		}
		c.JSON(200, out)
}
func handleAlbumAdd(c any) {
	var userId string
	var able map[string]any
	var name string
	var id string
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		name = OSLtoString(c.Param("name"))
		id = OSLtoString(c.Query("id"))
		if OSLequal(name, "") || OSLequal(id, "") {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "missing params",
				})
				return
		}
		addImageToAlbum(userId, name, id)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleAlbumRemove(c any) {
	var userId string
	var able map[string]any
	var name string
	var id string
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		name = OSLtoString(c.Param("name"))
		id = OSLtoString(c.Query("id"))
		if OSLequal(name, "") || OSLequal(id, "") {
				c.JSON(400, map[string]any{
						"ok": false,
						"error": "missing params",
				})
				return
		}
		removeImageFromAlbum(userId, name, id)
		c.JSON(200, map[string]any{
				"ok": true,
		})
}
func handleBinPreview(c any) {
	var userId string
	var able map[string]any
	var id string
	var path string
		userId = OSLtoString(c.MustGet("userId"))
		able = getAble(userId)
		if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
				c.JSON(401, map[string]any{
						"ok": false,
						"error": "not authorized",
				})
				return
		}
		id = OSLtoString(c.Param("id"))
		path = (((("db/" + OSLtoString(userId)) + "/bin/") + OSLtoString(id)) + ".jpg")
		servePreview(c, path, id, userId)
}
func handleShareInfo(c any) {
	var userId string
	var imageId string
	var sharesObj map[string]any
	var shares []any
	var share map[string]any
	var sharedWith []any
	var sharedWithInfo []any
	var targetUserId string
	var targetUsername string
	var profile map[string]any
	var data string
		userId = OSLtoString(c.MustGet("userId"))
		imageId = OSLtoString(c.Param("id"))
		sharesObj = readUserShares(userId)
		shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
		for i := 1; i <= OSLlen(shares); i++ {
				share = OSLcastObject(OSLgetItem(shares, i))
				if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(share, "imageId")), imageId)) {
						sharedWith = OSLcastArray(OSLgetItem(share, "sharedWith"))
						sharedWithInfo = []any{}
						for j := 1; j <= OSLlen(sharedWith); j++ {
								targetUserId = OSLtoString(OSLgetItem(sharedWith, j))
								targetUsername = ""
								if OSLcontains(userIdToUsername, targetUserId) {
										targetUsername = OSLtoString(OSLgetItem(userIdToUsername, targetUserId))
								} else if OSLcontains(userData, targetUserId) {
										profile = OSLcastObject(OSLgetItem(userData, targetUserId))
										if OSLcastBool(OSLnotEqual(OSLgetItem(profile, "username"), nil)) {
												targetUsername = OSLtoString(OSLgetItem(profile, "username"))
										}
								} else if fs.Exists((("db/" + OSLtoString(targetUserId)) + "/user.json")) {
										data = OSLtoString(fs.ReadFile((("db/" + OSLtoString(targetUserId)) + "/user.json")))
										profile = OSLcastObject(JsonParse(data))
										if OSLcastBool(OSLnotEqual(OSLgetItem(profile, "username"), nil)) {
												targetUsername = OSLtoString(OSLgetItem(profile, "username"))
												OSLsetItem(userData, targetUserId, profile)
												OSLsetItem(userIdToUsername, targetUserId, targetUsername)
										}
								}
								OSLappend(&(sharedWithInfo), map[string]any{
										"userId": targetUserId,
										"username": targetUsername,
								})
						}
						c.JSON(200, map[string]any{
								"ok": true,
								"sharedWith": sharedWithInfo,
								"isPublic": OSLgetItem(share, "isPublic"),
								"ownerId": userId,
						})
						return
				}
		}
		c.JSON(200, map[string]any{
				"ok": true,
				"sharedWith": []any{},
				"ownerId": userId,
		})
}
func handleDirectImage(c any) {
	var ownerUsername string
	var id string
	var sessionId string
	var requestingUserId string
	var ownerId string
	var allowed bool
	var sharesObj map[string]any
	var shares []any
	var share map[string]any
	var sharedWith []any
	var path string
		ownerUsername = OSLtoString(c.Param("username"))
		id = OSLtoString(c.Param("id"))
		sessionId = OSLtoString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
		requestingUserId = ""
		if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
				requestingUserId = OSLtoString(OSLgetItem(sessions, sessionId))
		}
		ownerId = getUserIdFromUsername(ownerUsername)
		if OSLcastBool(OSLequal(ownerId, "")) {
				if OSLcastBool(OSLequal(OSLlen(ownerUsername), 36)) {
						ownerId = ownerUsername
				}
				if OSLcastBool(OSLequal(ownerId, "")) {
						if OSLcastBool(OSLequal(requestingUserId, "")) {
								c.Redirect(302, "/auth")
								return
						}
						c.HTML(404, "error.html", map[string]any{
								"Title": "Not Found",
								"Message": "User not found.",
						})
						return
				}
		}
		allowed = OSLequal(ownerId, requestingUserId)
		if (allowed != true) {
				sharesObj = readUserShares(ownerId)
				shares = OSLcastArray(OSLgetItem(sharesObj, "shares"))
				for i := 1; i <= OSLlen(shares); i++ {
						share = OSLcastObject(OSLgetItem(shares, i))
						if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(share, "imageId")), id)) {
								if OSLcastBool(OSLequal(OSLgetItem(share, "isPublic"), true)) {
										allowed = true
										break
								}
								sharedWith = OSLcastArray(OSLgetItem(share, "sharedWith"))
								for j := 1; j <= OSLlen(sharedWith); j++ {
										if OSLcastBool(OSLequal(OSLtoString(OSLgetItem(sharedWith, j)), requestingUserId)) {
												allowed = true
												break
										}
								}
								if OSLcastBool(allowed) {
										break
								}
						}
				}
				if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
						c.Status(404)
						return
				}
				if OSLcastBool(OSLequal(requestingUserId, "")) {
						c.Redirect(302, "/auth")
						return
				}
				c.HTML(403, "error.html", map[string]any{
						"Title": "Access Denied",
						"Message": "You don't have access to this image.",
				})
				return
		}
		path = (((("db/" + OSLtoString(ownerId)) + "/blob/") + OSLtoString(id)) + ".jpg")
		if (fs.Exists(path) != true) || fs.IsDir(path) {
				if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
						c.Status(404)
						return
				}
				c.HTML(404, "error.html", map[string]any{
						"Title": "Not Found",
						"Message": "Image not found.",
				})
				return
		}
		if OSLcastBool(OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "HEAD")) {
				c.Header("Content-Type", "image/jpeg")
				c.Header("Cache-Control", "private, must-revalidate")
				c.Status(200)
				return
		}
		c.File(path)
}


var sessions map[string]any = map[string]any{}
var userData map[string]any = map[string]any{}
var userIdToUsername map[string]any = map[string]any{}
var usernameToUserId map[string]any = map[string]any{}
var useSubscriptions bool = false
var subscriptionSizes map[string]any = map[string]any{}
var quotas map[string]any = map[string]any{}
var downscaleWhen map[string]any = map[string]any{}
var authKey string = ""
func up(c any) {
		c.String(200, "ok")
}
func main() {
	var data string
		if OSLcastBool(fs.Exists("db/sessions.json")) {
				data = OSLtoString(fs.ReadFile("db/sessions.json"))
				sessions = JsonParse(data).(map[string]any)
		}
		loadConfig()
		var r = gin.Default()
		r.Use(noCORS)
		r.LoadHTMLGlob("templates/*")
		r.Static("/static", "./static")
		r.GET("/", requireSession, homePage)
		r.GET("/auth", authPage)
		var api = r.Group("/api")
		api.OPTIONS("/*cors", (func() {
				c.Status(200)
		}))
		api.GET("/auth", handleAuth)
		api.GET("/logout", handleLogout)
		api.GET("/able", requireSession, handleAble)
		api.GET("/storage", requireSession, handleStorage)
		api.GET("/search", requireSession, handleSearch)
		api.GET("/share/mine", requireSession, handleShareMine)
		api.GET("/share/others", requireSession, handleShareOthers)
		api.GET("/share/info/:id", requireSession, handleShareInfo)
		api.PATCH("/share/:id", requireSession, handleSharePatch)
		api.POST("/share", requireSession, handleShareCreate)
		api.GET("/shared/:owner/:id", handleSharedImage)
		api.GET("/shared/info/:owner/:id", handleSharedImageInfo)
		var images = api.Group("/images")
		images.GET("/all", requireSession, handleAllImages)
		images.GET("/recent", requireSession, handleRecentImages)
		images.GET("/:year", requireSession, handleYearImages)
		images.GET("/:year/:month", requireSession, handleMonthImages)
		images.DELETE("delete", requireSession, handleDeleteImages)
		var image = api.Group("/image")
		image.POST("", requireSession, handleUpload)
		image.POST("/upload", requireSession, handleUpload)
		image.DELETE("/:id", requireSession, handleDeleteImage)
		image.GET("/:id", requireSession, handleId)
		image.GET("/:id/info", requireSession, handleImageInfo)
		image.POST("/:id/rotate", requireSession, handleImageRotate)
		image.POST("/:id/compress", requireSession, handleImageCompress)
		image.POST("/:id/resize", requireSession, handleImageResize)
		image.GET("/preview/:id", requireSession, handlePreview)
		var albums = api.Group("/albums")
		albums.GET("", requireSession, handleAlbums)
		albums.POST("", requireSession, handleAlbumCreate)
		albums.DELETE("/:name", requireSession, handleAlbumDelete)
		albums.GET("/:name", requireSession, handleAlbumImages)
		albums.POST("/:name/add", requireSession, handleAlbumAdd)
		albums.POST("/:name/remove", requireSession, handleAlbumRemove)
		var bin = api.Group("/bin")
		bin.GET("", requireSession, handleBinList)
		bin.POST("/restore/:id", requireSession, handleBinRestore)
		bin.DELETE("/:id", requireSession, handleBinDelete)
		bin.POST("/empty", requireSession, handleBinEmpty)
		bin.GET("/preview/:id", requireSession, handleBinPreview)
		r.GET("/:username/:id", handleDirectImage)
		r.HEAD("/:username/:id", handleDirectImage)
		r.OPTIONS("/:username/:id", (func() any{
				c.Status(200)
		}))
		r.Run(":5607")
}


