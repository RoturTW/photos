package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"bytes"
	"encoding/json"
	"bufio"
	"os"
	"reflect"
	"io"
	"time"
	"math"
	"runtime"
	"sort"
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/rwcarlsen/goexif/exif"
	"path/filepath"
	OSL_bytes "bytes"
	OSL_image "image"
	"image/png"
	"image/jpeg"
	OSL_draw "golang.org/x/image/draw"
	OSL_img_resize "github.com/nfnt/resize"
	OSL_exif "github.com/rwcarlsen/goexif/exif"
	OSL_color "image/color"
	"net/http"
)

var wincreatetime float64 = OSLcastNumber(time.Now().UnixMilli())
var system_os = runtime.GOOS

// This is a set of funtions that are used in the compiler for OSL.go

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
	case []io.Reader:
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

func OSLcastString(s any) string {
	switch s := s.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	case []any:
		return JsonStringify(s)
	case map[string]any, map[string]string, map[string]int, map[string]float64, map[string]bool:
		return JsonStringify(s)
	case io.Reader:
		data, err := io.ReadAll(s)
		if err != nil {
			panic("OSLcastString: failed to read io.Reader:" + err.Error())
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
	switch s := s.(type) {
	case map[string]any:
		return s
	default:
		panic("OSLcastObject: invalid type, " + reflect.TypeOf(s).String())
	}
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
	return strings.EqualFold(OSLcastString(a), OSLcastString(b))
}

func OSLnotEqual(a any, b any) bool {
	if a == b {
		return false
	}
	return !strings.EqualFold(OSLcastString(a), OSLcastString(b))
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

func OSLsort(arr []any) []any {
	if arr == nil {
		return nil
	}

	sort.Slice(arr, func(i, j int) bool {
		return OSLcastString(arr[i]) < OSLcastString(arr[j])
	})
	return arr
}

func OSLsortBy(arr []any, key string) []any {
	if arr == nil {
		return nil
	}

	sort.Slice(arr, func(i, j int) bool {
		ai, ok1 := arr[i].(map[string]any)
		aj, ok2 := arr[j].(map[string]any)

		if !ok1 || !ok2 {
			return false
		}

		vi, _ := ai[key]
		vj, _ := aj[key]

		return OSLcastString(vi) < OSLcastString(vj)
	})
	return arr
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
		return T(rand.Intn(int(high-low)) + int(low))

	case float64:
		return (T(rand.Float64()) * (high - low)) + low
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

	if v, ok := a.(map[string]any); ok {
		return v[OSLcastString(b)]
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	key := OSLcastString(b)

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

	return any(OSLcastString(a) + OSLcastString(b)).(T)
}

func OSLadd[T float64 | int](a T, b T) T {
	return T(OSLcastNumber(a) + OSLcastNumber(b))
}

func OSLsub[T float64 | int](a T, b T) T {
	return T(OSLcastNumber(a) - OSLcastNumber(b))
}

func OSLmultiply[BT float64 | int](a any, b BT) any {
	if str, ok := a.(string); ok {
		n := OSLcastNumber(b)
		if n < 0 {
			return ""
		}
		return strings.Repeat(str, int(n))
	}

	return OSLcastNumber(a) * OSLcastNumber(b)
}

func OSLdivide[T float64 | int](a T, b T) T {
	return T(OSLcastNumber(a) / OSLcastNumber(b))
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

func OSLtrim(s any, from int, to int) string {
	str := []rune(OSLcastString(s))

	start := from - 1
	end := to

	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = len(str) + end + 1
	}

	if start > len(str) {
		start = len(str)
	}
	if end > len(str) {
		end = len(str)
	}

	if start > end {
		start, end = end, start
	}

	return string(str[start:end])
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

	key := OSLcastString(b)
	switch a := a.(type) {
	case map[string]any:
		_, ok := a[key]
		return ok
	case []any:
		for _, v := range a {
			if OSLcastString(v) == key {
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

	switch a := a.(type) {
	case map[string]any:
		delete(a, OSLcastString(b))
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

	key := OSLcastString(b)

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

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	key := OSLcastString(b)

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
		if field.IsValid() && field.CanSet() {
			val := reflect.ValueOf(value)
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return true
			}
		}
		return false
	}

	return false
}

func OSLarrayJoin(a any, b any) string {
	var out string
	sep := OSLcastString(b)
	arr := OSLcastArray(a)

	for _, v := range arr {
		out += OSLcastString(v) + sep
	}

	return strings.TrimSuffix(out, sep)
}

func OSLgetKeys(a any) []any {
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
		i := 0
		for _, v := range a {
			keys[i] = OSLcastString(v)
			i++
		}
		return keys
	default:
		return []any{}
	}
}

func OSLgetValues(a any) []any {
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
	switch a := a.(type) {
	case map[string]any:
		_, ok := a[OSLcastString(b)]
		return ok
	case []any:
		for _, v := range a {
			if OSLcastString(v) == OSLcastString(b) {
				return true
			}
		}
		return false
	case string:
		return strings.Contains(a, OSLcastString(b))
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


// name: requests
// description: HTTP utilities
// author: Mist
// requires: net/http, encoding/json, io

type HTTP struct {
	Client *http.Client
}

func extractHeadersAndBody(data map[string]any) (headers map[string]string, body io.Reader) {
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
			headers[k] = OSLcastString(v)
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

	respBody, err := io.ReadAll(resp.Body)
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
	return h.doRequest(http.MethodGet, OSLcastString(url), m)
}

func (h *HTTP) Post(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPost, OSLcastString(url), data)
}

func (h *HTTP) Put(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPut, OSLcastString(url), data)
}

func (h *HTTP) Patch(url any, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPatch, OSLcastString(url), data)
}

func (h *HTTP) Delete(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodDelete, OSLcastString(url), m)
}

func (h *HTTP) Options(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodOptions, OSLcastString(url), m)
}

func (h *HTTP) Head(url any, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	out := map[string]any{"success": false}
	headers, _ := extractHeadersAndBody(m)
	req, err := http.NewRequest(http.MethodHead, OSLcastString(url), nil)
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

func (IMG) Decode(r io.Reader) *OSL_img_Image {
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

func (IMG) DecodeSize(r io.Reader) (int, int) {
	cfg, _, err := OSL_image.DecodeConfig(r)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

/* -------------------- encode -------------------- */

func (IMG) EncodePNG(w io.Writer, i *OSL_img_Image) bool {
	return i != nil && !i.closed && png.Encode(w, i.im) == nil
}

func (IMG) EncodeJPEG(w io.Writer, i *OSL_img_Image, q int) bool {
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

func (IMG) NormalizeOrientation(i *OSL_img_Image, r io.Reader) *OSL_img_Image {
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
// name: fs
// description: File system utilities
// author: Mist
// requires: os, path/filepath

type FS struct{}

func (FS) ReadFile(path any) string {
	data, err := os.ReadFile(OSLcastString(path))
	if err != nil {
		return ""
	}
	return string(data)
}

func (FS) ReadFileBytes(path any) []byte {
	data, err := os.ReadFile(OSLcastString(path))
	if err != nil {
		return []byte{}
	}
	return data
}

func (FS) WriteFile(path any, data any) bool {
	err := os.WriteFile(OSLcastString(path), []byte(OSLcastString(data)), 0644)
	return err == nil
}

func (FS) Rename(oldPath any, newPath any) bool {
	err := os.Rename(OSLcastString(oldPath), OSLcastString(newPath))
	return err == nil
}

func (FS) Exists(path any) bool {
	_, err := os.Stat(OSLcastString(path))
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
	err := os.Mkdir(OSLcastString(path), 0755)
	return err == nil
}

func (FS) MkdirAll(path any) bool {
	err := os.MkdirAll(OSLcastString(path), 0755)
	return err == nil
}

func (FS) CopyDir(srcPath any, dstPath any) bool {
	src := OSLcastString(srcPath)
	dst := OSLcastString(dstPath)

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

		if _, err := io.Copy(out, in); err != nil {
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
	files, err := os.ReadDir(OSLcastString(path))
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
	dir := OSLcastString(path)
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
	dir := OSLcastString(path)
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
	info, err := os.Stat(OSLcastString(path))
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
	err := os.Chdir(OSLcastString(path))
	return err == nil
}

func (FS) JoinPath(path ...any) string {
	stringPath := make([]string, len(path))
	for i, p := range path {
		stringPath[i] = OSLcastString(p)
	}
	return filepath.Join(stringPath...)
}

func (FS) GetBase(path any) string {
	return filepath.Base(OSLcastString(path))
}

func (FS) GetDir(path any) string {
	return filepath.Dir(OSLcastString(path))
}

func (FS) GetExt(path any) string {
	return filepath.Ext(OSLcastString(path))
}

func (FS) GetParts(path any) []any {
	stringPath := OSLcastString(path)
	return []any{filepath.Base(stringPath), filepath.Dir(stringPath), filepath.Ext(stringPath)}
}

func (FS) GetSize(path any) float64 {
	info, err := os.Stat(OSLcastString(path))
	if err != nil {
		return 0
	}
	return float64(info.Size())
}

func (FS) GetModTime(path any) float64 {
	info, err := os.Stat(OSLcastString(path))
	if err != nil {
		return 0.0
	}
	return float64(info.ModTime().UnixMilli())
}

func (FS) GetStat(path any) map[string]any {
	info, err := os.Stat(OSLcastString(path))
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
	pathStr := OSLcastString(path)
	absPath, err := filepath.EvalSymlinks(pathStr)
	if err != nil {
		return ""
	}
	return absPath
}

// Global instance
var fs = FS{}
func handleAuth(c *gin.Context) {
  var token string = OSLcastString(c.Query("v"))
  if OSLequal(token, "") {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "missing token",
    })
    return
  }
  var resp map[string]any = OSLcastObject(requests.Get(("https://api.rotur.dev/validate?key=rotur-photos&v=" + token)))
  if (OSLgetItem(resp, "success") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "failed to validate token",
    })
    return
  }
  var json map[string]any = OSLcastObject(JsonParse(OSLcastString(OSLgetItem(resp, "body"))))
  if OSLnotEqual(OSLgetItem(json, "error"), nil) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": OSLgetItem(json, "error"),
    })
    return
  }
  if (OSLgetItem(json, "valid") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "invalidate token",
    })
    return
  }
  var sessionId string = OSLcastString(randomString(32))
  var username string = OSLcastString(OSLgetItem(OSLSplit(token, ","), 1))
  OSLsetItem(sessions, sessionId, username)
  var profileReq map[string]any = OSLcastObject(writeProfile(username))
  if (OSLgetItem(profileReq, "ok") != true) {
    c.JSON(401, OSLgetItem(profileReq, "error"))
    return
  }
  fs.WriteFile("db/sessions.json", JsonFormat(sessions))
  c.SetCookie("session_id", sessionId, 86400, "/", "", false, true)
  c.JSON(200,  map[string]any{
    "ok": true,
    "token": token,
  })
}
func handleLogout(c *gin.Context) {
  var sessionId string = OSLcastString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
  if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
    OSLdelete(sessions, sessionId)
    fs.WriteFile("db/sessions.json", JsonFormat(sessions))
  }
  c.SetCookie("session_id", "", -1, "/", "", false, true)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleAble(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  c.JSON(200, able)
}
func handleStorage(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var stats map[string]any = OSLcastObject(calculateStorageStats(username))
  OSLsetItem(stats, "quota", OSLgetItem(able, "storageQuota"))
  c.JSON(200, stats)
}
func handleAllImages(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var images []any = OSLcastArray(readUserImages(username))
  images = enrichImagesWithSharing(username, images)
  c.JSON(200, images)
}
func handleRecentImages(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var arr []any = OSLcastArray(readUserImages(username))
  var n int = OSLlen(arr)
  var nowMs float64 = OSLcastNumber(time.Now().UnixMilli())
  var ninety float64 = OSLcastNumber(OSLmultiply(OSLmultiply(OSLmultiply(OSLmultiply(90, 24), 60), 60), 1000))
  var out []any = []any{}
  for i := 1; i <= OSLround(n); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    var ts float64 = OSLcastNumber(OSLgetItem(it, "timestamp"))
    if OSLcastNumber(OSLsub(nowMs, ts)) <= OSLcastNumber(ninety) {
      OSLappend(&(out), it)
    }
  }
  out = enrichImagesWithSharing(username, out)
  c.JSON(200, out)
}
func handleYearImages(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var year int = OSLcastInt(c.Param("year"))
  var arr []any = OSLcastArray(readUserImages(username))
  var n int = OSLlen(arr)
  var out []any = []any{}
  for i := 1; i <= OSLround(n); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    var ts int = OSLcastInt(OSLgetItem(it, "timestamp"))
    var t = time.UnixMilli(int64(ts))
    if OSLequal(t.Year(), year) {
      OSLappend(&(out), it)
    }
  }
  out = enrichImagesWithSharing(username, out)
  c.JSON(200, out)
}
func handleMonthImages(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var year int = OSLcastInt(c.Param("year"))
  var month int = OSLcastInt(c.Param("month"))
  var arr []any = OSLcastArray(readUserImages(username))
  var n int = OSLlen(arr)
  var out []any = []any{}
  for i := 1; i <= OSLround(n); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    var ts int = OSLcastInt(OSLgetItem(it, "timestamp"))
    var t = time.UnixMilli(int64(ts))
    var tMonth int = int(0)
    tMonth = int(t.Month())
    if OSLequal(t.Year(), year) && OSLequal(tMonth, month) {
      OSLappend(&(out), it)
    }
  }
  out = enrichImagesWithSharing(username, out)
  c.JSON(200, out)
}
func handleUpload(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var base string = (("db/" + strings.ToLower(username)) + "/blob")
  if (fs.Exists(base) != true) {
    fs.MkdirAll(base)
  }
  var tmpPath string = ((base + "/tmp-") + randomString(16))
  var body []byte = OSLgetItem(OSLcastArray(c.GetRawData()), 1).([]byte)
  if OSLequal(OSLlen(body), 0) {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "empty body",
    })
    return
  }
  var stats map[string]any = OSLcastObject(calculateStorageStats(username))
  if OSLcastNumber(((OSLcastNumber(OSLgetItem(stats, "totalBytes")) + OSLcastNumber(OSLgetItem(stats, "binBytes"))) + float64(OSLlen(body)))) > OSLcastNumber(OSLcastNumber(OSLgetItem(able, "storageQuota"))) {
    c.JSON(403,  map[string]any{
      "ok": false,
      "error": "storage quota exceeded",
    })
    return
  }
  if (fs.WriteFile(tmpPath, body) != true) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "failed to write upload",
    })
    return
  }
  var im = img.Open(tmpPath)
  if OSLequal(im, nil) {
    fs.Remove(tmpPath)
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid image",
    })
    return
  }
  var size map[string]any = OSLcastObject(im.Size())
  var w int = OSLround(OSLgetItem(size, "w"))
  var h int = OSLround(OSLgetItem(size, "h"))
  if OSLcastNumber(w) > OSLcastNumber(OSLgetItem(downscaleWhen, "width")) || OSLcastNumber(h) > OSLcastNumber(OSLgetItem(downscaleWhen, "height")) {
    var ratio float64 = OSLcastNumber(OSLmin(OSLdivide(OSLcastInt(OSLgetItem(downscaleWhen, "width")), w), OSLdivide(OSLcastInt(OSLgetItem(downscaleWhen, "height")), h)))
    w = OSLround(OSLmultiply(w, ratio))
    h = OSLround(OSLmultiply(h, ratio))
    var resized = img.Resize(im, w, h)
    im.Close()
    if OSLequal(resized, nil) {
      fs.Remove(tmpPath)
      c.JSON(500,  map[string]any{
        "ok": false,
        "error": "resize failed",
      })
      return
    }
    im = resized
  }
  var id string = OSLcastString(randomString(24))
  var outPath string = (((base + "/") + id) + ".jpg")
  if (img.SaveJPEG(im, outPath, 90) != true) {
    im.Close()
    fs.Remove(tmpPath)
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "save failed",
    })
    return
  }
  im.Close()
  var tsms float64 = OSLcastNumber(time.Now().UnixMilli())
  var x *exif.Exif = nil
  f, err := os.Open(tmpPath)
  if OSLequal(err, nil) {
    defer f.Close()
    x, err = exif.Decode(f)
    if OSLequal(err, nil) {
      t, err := x.DateTime()
      if OSLequal(err, nil) && (t.IsZero() != true) {
        tsms = OSLcastNumber(t.UnixMilli())
      }
    }
  }
  var entry map[string]any =  map[string]any{
    "id": id,
    "width": w,
    "height": h,
    "timestamp": tsms,
    "make": "",
    "model": "",
    "exposure_time": "",
    "f_number": "",
  }
  if OSLnotEqual(x, nil) {
    if tag, err := x.Get(exif.Make); err == nil { entry["make"], _ = tag.StringVal() }
    if tag, err := x.Get(exif.Model); err == nil { entry["model"], _ = tag.StringVal() }
    if tag, err := x.Get(exif.ExposureTime); err == nil { entry["exposure_time"], _ = tag.StringVal() }
    if tag, err := x.Get(exif.FNumber); err == nil { entry["f_number"], _ = tag.StringVal() }
  }
  fs.Remove(tmpPath)
  var images []any = OSLcastArray(readUserImages(username))
  OSLappend(&(images), entry)
  writeUserImages(username, images)
  c.JSON(200,  map[string]any{
    "ok": true,
    "id": id,
  })
}
func handleId(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var json bool = false
  if strings.HasSuffix(id, ".json") {
    id = OSLtrim(id, 1, -6)
    json = true
  }
  if json {
    var it map[string]any = OSLcastObject(findImage(readUserImages(username), id))
    if OSLequal(OSLgetItem(it, "id"), nil) {
      c.JSON(404,  map[string]any{
        "ok": false,
        "error": "not found",
      })
      return
    }
    c.JSON(200, it)
    return
  }
  var path string = (((("db/" + strings.ToLower(username)) + "/blob/") + id) + ".jpg")
  if (fs.Exists(path) != true) || fs.IsDir(path) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  var fileInfo map[string]any = OSLcastObject(fs.GetStat(path))
  var etag string = (("\"" + OSLcastString(OSLgetItem(fileInfo, "modTime"))) + "\"")
  var ifNoneMatch string = OSLcastString(c.GetHeader("If-None-Match"))
  if OSLequal(ifNoneMatch, etag) {
    c.Status(304)
    return
  }
  c.Header("ETag", etag)
  c.Header("Cache-Control", "private, must-revalidate")
  c.File(path)
}
func handleImageRotate(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var body map[string]any = OSLcastObject(JsonParse(OSLcastString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
  var angle float64 = OSLcastNumber(OSLgetItem(body, "angle"))
  var path string = (((("db/" + strings.ToLower(username)) + "/blob/") + id) + ".jpg")
  if (fs.Exists(path) != true) || fs.IsDir(path) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  var imgFile = img.Open(path)
  if OSLequal(imgFile, nil) {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid image",
    })
    return
  }
  defer imgFile.Close()
  var rotated = img.Rotate(imgFile, angle)
  if OSLequal(rotated, nil) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "failed to rotate",
    })
    return
  }
  defer rotated.Close()
  var wrote bool = bool(img.SaveJPEG(rotated, path, 90))
  if (wrote != true) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "failed to save",
    })
    return
  }
  var arr []any = OSLcastArray(readUserImages(username))
  var it map[string]any = OSLcastObject(findImage(arr, id))
  if OSLnotEqual(OSLgetItem(it, "id"), nil) {
    if OSLequal(angle, 90) || OSLequal(angle, 270) {
      var oldW float64 = OSLcastNumber(OSLgetItem(it, "width"))
      OSLsetItem(it, "width", OSLgetItem(it, "height"))
      OSLsetItem(it, "height", oldW)
      writeUserImages(username, arr)
    }
  }
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleSearch(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var q string = strings.ToLower(OSLcastString(c.Query("q")))
  var arr []any = OSLcastArray(readUserImages(username))
  if OSLequal(q, "") {
    c.JSON(200, arr)
    return
  }
  var months []any = []any{
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
  var albums map[string]any = OSLcastObject(readUserAlbums(username))
  var out []any = []any{}
  var n int = OSLlen(arr)
  for i := 1; i <= OSLround(n); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    var match bool = false
    var ts float64 = OSLcastNumber(OSLgetItem(it, "timestamp"))
    if OSLcastNumber(ts) > OSLcastNumber(0) {
      var t = time.UnixMilli(int64(ts))
      var year int = OSLcastInt(t.Year())
      var month int = OSLcastInt(int(t.Month()))
      if OSLcontains(q, OSLcastString(year)) {
        match = true
      }
      if OSLcastNumber(month) > OSLcastNumber(0) && OSLcastNumber(month) <= OSLcastNumber(12) {
        var monthName string = OSLcastString(OSLgetItem(months, month))
        if OSLcontains(q, monthName) {
          match = true
        }
      }
    }
    var make string = strings.ToLower(OSLcastString(OSLgetItem(it, "make")))
    var model string = strings.ToLower(OSLcastString(OSLgetItem(it, "model")))
    if OSLnotEqual(make, "") && OSLcontains(q, make) {
      match = true
    }
    if OSLnotEqual(model, "") && OSLcontains(q, model) {
      match = true
    }
    var items map[string]any = OSLcastObject(OSLgetItem(albums, "items"))
    for j := 1; j <= OSLlen(OSLgetItem(albums, "albums")); j++ {
      var albumName string = OSLcastString(OSLgetItem(OSLgetItem(albums, "albums"), j))
      if OSLcontains(strings.ToLower(albumName), q) {
        var albumIds []any = OSLcastArray(OSLgetItem(items, albumName))
        for k := 1; k <= OSLlen(albumIds); k++ {
          if OSLequal(OSLcastString(OSLgetItem(albumIds, k)), OSLcastString(OSLgetItem(it, "id"))) {
            match = true
          }
        }
      }
    }
    if match {
      OSLappend(&(out), it)
    }
  }
  c.JSON(200, out)
}
func handleShareMine(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var sharesObj map[string]any = OSLcastObject(readUserShares(username))
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  var images []any = OSLcastArray(readUserImages(username))
  var out []any = []any{}
  for i := 1; i <= OSLlen(shares); i++ {
    var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
    var imageId string = OSLcastString(OSLgetItem(share, "imageId"))
    var img map[string]any = OSLcastObject(findImage(images, imageId))
    if OSLnotEqual(OSLgetItem(img, "id"), nil) {
      OSLsetItem(img, "sharedWith", OSLgetItem(share, "sharedWith"))
      OSLappend(&(out), img)
    }
  }
  c.JSON(200, out)
}
func handleShareOthers(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var sharedWithMe []any = OSLcastArray(getSharedWithMe(username))
  var results []any = []any{}
  for i := 1; i <= OSLlen(sharedWithMe); i++ {
    var item map[string]any = OSLcastObject(OSLgetItem(sharedWithMe, i))
    var owner string = OSLcastString(OSLgetItem(item, "owner"))
    var imageId string = OSLcastString(OSLgetItem(item, "imageId"))
    var ownerImages []any = OSLcastArray(readUserImages(owner))
    var img map[string]any = OSLcastObject(findImage(ownerImages, imageId))
    if OSLnotEqual(OSLgetItem(img, "id"), nil) {
      OSLsetItem(img, "owner", owner)
      OSLappend(&(results), img)
    }
  }
  c.JSON(200, results)
}
func handleSharePatch(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var imageId string = OSLcastString(c.Param("id"))
  var body map[string]any = OSLcastObject(JsonParse(OSLcastString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
  var add []any = OSLcastArray(OSLgetItem(body, "add"))
  var remove []any = OSLcastArray(OSLgetItem(body, "remove"))
  if OSLequal(add, nil) {
    add = []any{}
  }
  if OSLequal(remove, nil) {
    remove = []any{}
  }
  for i := 1; i <= OSLlen(add); i++ {
    addShare(username, imageId, OSLcastString(OSLgetItem(add, i)))
  }
  for i := 1; i <= OSLlen(remove); i++ {
    removeShare(username, imageId, OSLcastString(OSLgetItem(remove, i)))
  }
  if OSLnotEqual(OSLgetItem(body, "isPublic"), nil) {
    setPublicShare(username, imageId, OSLcastBool(OSLgetItem(body, "isPublic")))
  }
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleShareCreate(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var body map[string]any = OSLcastObject(JsonParse(OSLcastString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))))
  var imageId string = OSLcastString(OSLgetItem(body, "imageId"))
  var targetUser string = OSLcastString(OSLgetItem(body, "username"))
  if OSLequal(imageId, "") || OSLequal(targetUser, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "missing imageId or username",
    })
    return
  }
  var images []any = OSLcastArray(readUserImages(username))
  var img map[string]any = OSLcastObject(findImage(images, imageId))
  if OSLequal(OSLgetItem(img, "id"), nil) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "image not found",
    })
    return
  }
  addShare(username, imageId, targetUser)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleSharedImage(c *gin.Context) {
  var owner string = OSLcastString(c.Param("owner"))
  var id string = OSLcastString(c.Param("id"))
  var requestingUser string = ""
  var sessionId string = OSLcastString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
  if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
    requestingUser = strings.ToLower(OSLcastString(OSLgetItem(sessions, sessionId)))
  }
  var sharesObj map[string]any = OSLcastObject(readUserShares(owner))
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  var allowed bool = false
  if OSLequal(requestingUser, strings.ToLower(owner)) {
    allowed = true
  }
  for i := 1; i <= OSLlen(shares); i++ {
    var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
    if OSLequal(OSLcastString(OSLgetItem(share, "imageId")), id) {
      var sharedWith []any = OSLcastArray(OSLgetItem(share, "sharedWith"))
      for j := 1; j <= OSLlen(sharedWith); j++ {
        if OSLequal(strings.ToLower(OSLcastString(OSLgetItem(sharedWith, j))), requestingUser) {
          allowed = true
        }
      }
    }
  }
  if (allowed != true) {
    c.JSON(403,  map[string]any{
      "ok": false,
      "error": "not authorized to view this image",
    })
    return
  }
  var path string = (((("db/" + strings.ToLower(owner)) + "/blob/") + id) + ".jpg")
  if (fs.Exists(path) != true) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  c.File(path)
}
func handleSharedImageInfo(c *gin.Context) {
  var owner string = OSLcastString(c.Param("owner"))
  var id string = OSLcastString(c.Param("id"))
  var requestingUser string = ""
  var sessionId string = OSLcastString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
  if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
    requestingUser = strings.ToLower(OSLcastString(OSLgetItem(sessions, sessionId)))
  }
  var sharesObj map[string]any = OSLcastObject(readUserShares(owner))
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  var allowed bool = false
  if OSLequal(requestingUser, strings.ToLower(owner)) {
    allowed = true
  }
  for i := 1; i <= OSLlen(shares); i++ {
    var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
    if OSLequal(OSLcastString(OSLgetItem(share, "imageId")), id) {
      var sharedWith []any = OSLcastArray(OSLgetItem(share, "sharedWith"))
      for j := 1; j <= OSLlen(sharedWith); j++ {
        if OSLequal(strings.ToLower(OSLcastString(OSLgetItem(sharedWith, j))), requestingUser) {
          allowed = true
        }
      }
    }
  }
  if (allowed != true) {
    c.JSON(403,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var images []any = OSLcastArray(readUserImages(owner))
  var img map[string]any = OSLcastObject(findImage(images, id))
  if OSLequal(OSLgetItem(img, "id"), nil) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  OSLsetItem(img, "owner", owner)
  c.JSON(200, img)
}
func handlePreview(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var path string = (((("db/" + strings.ToLower(username)) + "/blob/") + id) + ".jpg")
  servePreview(c, path, id, username)
}
func handleDeleteImage(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var base string = ("db/" + strings.ToLower(username))
  var path string = (((base + "/blob/") + id) + ".jpg")
  var binBase string = (base + "/bin")
  if (fs.Exists(binBase) != true) {
    fs.MkdirAll(binBase)
  }
  var data []byte = fs.ReadFileBytes(path)
  if OSLcastNumber(OSLlen(data)) > OSLcastNumber(0) {
    var imgFile = img.DecodeBytes(data)
    if OSLnotEqual(imgFile, nil) {
      var s map[string]any = OSLcastObject(imgFile.Size())
      var w int = OSLround(OSLgetItem(s, "w"))
      var h int = OSLround(OSLgetItem(s, "h"))
      var resized = img.Resize(imgFile, w, h)
      if OSLequal(resized, nil) {
        c.JSON(500,  map[string]any{
          "ok": false,
          "error": "resize failed",
        })
        return
      }
      var binPath string = (((binBase + "/") + id) + ".jpg")
      var wrote bool = img.SaveJPEG(resized, binPath, 90)
      if wrote {
        var binArr []any = OSLcastArray(readUserBin(username))
        var tsms float64 = OSLcastNumber(time.Now().UnixMilli())
        var arr []any = OSLcastArray(readUserImages(username))
        var it map[string]any = OSLcastObject(findImage(arr, id))
        var origTs float64 = OSLcastNumber(OSLgetItem(it, "timestamp"))
        if OSLequal(origTs, 0) {
          origTs = tsms
        }
        var entry map[string]any =  map[string]any{
          "id": id,
          "width": w,
          "height": h,
          "timestamp": origTs,
          "deletedAt": tsms,
        }
        OSLappend(&(binArr), entry)
        writeUserBin(username, binArr)
      }
    }
  }
  if fs.Exists(path) {
    fs.Remove(path)
  }
  var arr []any = OSLcastArray(readUserImages(username))
  var out []any = OSLcastArray(removeImage(arr, id))
  writeUserImages(username, out)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleDeleteImages(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var ids any = OSLgetItem(JsonParse(OSLcastString(OSLgetItem(OSLgetItem(c, "request"), "body"))), "ids")
  if OSLequal(OSLlen(ids), 0) {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "missing ids",
    })
    return
  }
  if OSLnotEqual(OSLtypeof(ids), "array") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid ids",
    })
    return
  }
  var arr []any = OSLcastArray(readUserImages(username))
  var out []any = OSLcastArray(removeImages(arr, ids.([]any)))
  writeUserImages(username, out)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleBinList(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var arr []any = OSLcastArray(readUserBin(username))
  c.JSON(200, arr)
}
func handleBinRestore(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var base string = ("db/" + strings.ToLower(username))
  var binPath string = (((base + "/bin/") + id) + ".jpg")
  var blobPath string = (((base + "/blob/") + id) + ".jpg")
  var imgFile = img.Open(binPath)
  if OSLequal(imgFile, nil) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  defer imgFile.Close()
  var size map[string]any = OSLcastObject(imgFile.Size())
  var w int = OSLround(OSLgetItem(size, "w"))
  var h int = OSLround(OSLgetItem(size, "h"))
  var resized = img.Resize(imgFile, w, h)
  if OSLequal(resized, nil) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "failed to resize",
    })
    return
  }
  defer resized.Close()
  var wrote bool = img.SaveJPEG(resized, blobPath, 90)
  if (wrote != true) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "failed to restore",
    })
    return
  }
  var binArr []any = OSLcastArray(readUserBin(username))
  var newBin []any = []any{}
  var ts float64 = OSLcastNumber(time.Now().UnixMilli())
  for i := 1; i <= OSLlen(binArr); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(binArr, i))
    if OSLnotEqual(OSLcastString(OSLgetItem(it, "id")), id) {
      OSLappend(&(newBin), it)
    } else {
      ts = OSLcastNumber(OSLgetItem(it, "timestamp"))
    }
  }
  writeUserBin(username, newBin)
  var arr []any = OSLcastArray(readUserImages(username))
  var entry map[string]any =  map[string]any{
    "id": id,
    "width": w,
    "height": h,
    "timestamp": ts,
  }
  OSLappend(&(arr), entry)
  writeUserImages(username, arr)
  fs.Remove(binPath)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleBinDelete(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var base string = ("db/" + strings.ToLower(username))
  var binPath string = (((base + "/bin/") + id) + ".jpg")
  if fs.Exists(binPath) {
    fs.Remove(binPath)
  }
  var binArr []any = OSLcastArray(readUserBin(username))
  var newBin []any = []any{}
  for i := 1; i <= OSLlen(binArr); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(binArr, i))
    if OSLnotEqual(OSLcastString(OSLgetItem(it, "id")), id) {
      OSLappend(&(newBin), it)
    }
  }
  writeUserBin(username, newBin)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleBinEmpty(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  writeUserBin(username, []any{})
  var base string = ("db/" + strings.ToLower(username))
  var binDir string = (base + "/bin")
  fs.Remove(binDir)
  fs.Mkdir(binDir)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleAlbums(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var albums map[string]any = OSLcastObject(readUserAlbums(username))
  c.JSON(200, OSLgetItem(albums, "albums"))
}
func handleAlbumCreate(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var name string = OSLcastString(c.Query("name"))
  if OSLequal(name, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "missing name",
    })
    return
  }
  var albums map[string]any = OSLcastObject(addAlbum(username, name))
  c.JSON(200, OSLgetItem(albums, "albums"))
}
func handleAlbumDelete(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var name string = OSLcastString(c.Param("name"))
  var albums map[string]any = OSLcastObject(removeAlbumDef(username, name))
  c.JSON(200, OSLgetItem(albums, "albums"))
}
func handleAlbumImages(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var name string = OSLcastString(c.Param("name"))
  var albums map[string]any = OSLcastObject(readUserAlbums(username))
  var ids []any = OSLcastArray(OSLgetItem(OSLgetItem(albums, "items"), name))
  var all []any = OSLcastArray(readUserImages(username))
  var out []any = []any{}
  for i := 1; i <= OSLlen(ids); i++ {
    var id string = OSLcastString(OSLgetItem(ids, i))
    var it map[string]any = OSLcastObject(findImage(all, id))
    if OSLnotEqual(OSLgetItem(it, "id"), nil) {
      OSLappend(&(out), it)
    }
  }
  c.JSON(200, out)
}
func handleAlbumAdd(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var name string = OSLcastString(c.Param("name"))
  var id string = OSLcastString(c.Query("id"))
  if OSLequal(name, "") || OSLequal(id, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "missing params",
    })
    return
  }
  addImageToAlbum(username, name, id)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleAlbumRemove(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var name string = OSLcastString(c.Param("name"))
  var id string = OSLcastString(c.Query("id"))
  if OSLequal(name, "") || OSLequal(id, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "missing params",
    })
    return
  }
  removeImageFromAlbum(username, name, id)
  c.JSON(200,  map[string]any{
    "ok": true,
  })
}
func handleBinPreview(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) && (OSLgetItem(able, "hasImages") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var path string = (((("db/" + strings.ToLower(username)) + "/bin/") + id) + ".jpg")
  servePreview(c, path, id, username)
}
func handleShareInfo(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var imageId string = OSLcastString(c.Param("id"))
  var sharesObj map[string]any = OSLcastObject(readUserShares(username))
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  for i := 1; i <= OSLlen(shares); i++ {
    var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
    if OSLequal(OSLcastString(OSLgetItem(share, "imageId")), imageId) {
      c.JSON(200,  map[string]any{
        "ok": true,
        "sharedWith": OSLgetItem(share, "sharedWith"),
        "isPublic": OSLgetItem(share, "isPublic"),
      })
      return
    }
  }
  c.JSON(200,  map[string]any{
    "ok": true,
    "sharedWith": []any{},
  })
}
func handleDirectImage(c *gin.Context) {
  var owner string = OSLcastString(c.Param("username"))
  var id string = OSLcastString(c.Param("id"))
  var sessionId string = OSLcastString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
  var requestingUser string = ""
  if OSLnotEqual(sessionId, "") && OSLcontains(sessions, sessionId) {
    requestingUser = strings.ToLower(OSLcastString(OSLgetItem(sessions, sessionId)))
  }
  var allowed bool = OSLequal(strings.ToLower(owner), requestingUser)
  if (allowed != true) {
    var sharesObj map[string]any = OSLcastObject(readUserShares(owner))
    var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
    for i := 1; i <= OSLlen(shares); i++ {
      var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
      if OSLequal(OSLcastString(OSLgetItem(share, "imageId")), id) {
        if OSLequal(OSLgetItem(share, "isPublic"), true) {
          allowed = true
          break
        }
        var sharedWith []any = OSLcastArray(OSLgetItem(share, "sharedWith"))
        for j := 1; j <= OSLlen(sharedWith); j++ {
          if OSLequal(strings.ToLower(OSLcastString(OSLgetItem(sharedWith, j))), requestingUser) {
            allowed = true
            break
          }
        }
        if allowed {
          break
        }
      }
    }
  }
  if (allowed != true) {
    if OSLequal(requestingUser, "") {
      c.Redirect(302, "/auth")
      return
    }
    c.HTML(403, "error.html",  map[string]any{
      "Title": "Access Denied",
      "Message": "You don't have access to this image.",
    })
    return
  }
  var path string = (((("db/" + strings.ToLower(owner)) + "/blob/") + id) + ".jpg")
  if (fs.Exists(path) != true) || fs.IsDir(path) {
    if OSLequal(requestingUser, "") {
      c.Redirect(302, "/auth")
      return
    }
    c.HTML(404, "error.html",  map[string]any{
      "Title": "Not Found",
      "Message": "Image not found.",
    })
    return
  }
  c.File(path)
}

func abortUnauthorized(c *gin.Context) {
  var path string = OSLcastString(OSLgetItem(OSLgetItem(OSLgetItem(c, "Request"), "URL"), "Path"))
  if strings.HasPrefix(path, "/api") {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authenticated",
    })
  } else {
    c.Redirect(302, "/auth")
  }
  c.Abort()
}
func requireSession(c *gin.Context) {
  var sessionId string = OSLcastString(OSLgetItem(OSLcastArray(c.Cookie("session_id")), 1))
  if OSLequal(sessionId, "") {
    abortUnauthorized(c)
    return
  }
  if (OSLcontains(sessions, sessionId) != true) {
    abortUnauthorized(c)
    return
  }
  var username string = OSLcastString(OSLgetItem(sessions, sessionId))
  if (OSLcontains(userData, strings.ToLower(username)) != true) {
    if fs.Exists((("db/" + strings.ToLower(username)) + "/user.json")) {
      var data string = OSLcastString(fs.ReadFile((("db/" + strings.ToLower(username)) + "/user.json")))
      OSLsetItem(userData, strings.ToLower(username), JsonParse(data))
    }
  }
  c.Set("username", username)
  c.Next()
}

func homePage(c *gin.Context) {
  var username = OSLcastString(c.MustGet("username"))
  var profileReq map[string]any = OSLcastObject(writeProfile(username))
  if (OSLgetItem(profileReq, "ok") != true) {
    c.String(401, OSLcastString(OSLgetItem(profileReq, "error")))
    return
  }
  c.HTML(200, "index.html",  map[string]any{
    "Username": username,
    "Subscription": OSLgetItem(OSLgetItem(userData, strings.ToLower(username)), "subscription"),
  })
}
func authPage(c *gin.Context) {
  c.HTML(200, "auth.html",  map[string]any{
    "AuthKey": authKey,
  })
}

func randomString(length int) string {
  var chars string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
  var result string = ""
  for i := 1; i <= OSLround(length); i++ {
    result = (result + OSLcastString(OSLgetItem(chars, OSLadd(rand.Intn(OSLlen(chars)), 1))))
  }
  return result
}
func calculateFileHash(path string) string {
  var hash string = ""
  f, err := os.Open(path)
  if OSLequal(err, nil) {
    h := md5.New()
    if _, err := io.Copy(h, f); err == nil { hash = hex.EncodeToString(h.Sum(nil)) }
    f.Close()
  }
  return hash
}
func noCORS(c *gin.Context) {
  c.Header("Access-Control-Allow-Origin", "*")
  var origin string = OSLcastString(c.GetHeader("Origin"))
  if OSLnotEqual(origin, "") {
    c.Header("Access-Control-Allow-Origin", origin)
    c.Header("Access-Control-Allow-Credentials", "true")
    c.Next()
    return
  }
  c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
  c.Header("Access-Control-Allow-Headers", "*")
  if OSLequal(OSLgetItem(OSLgetItem(c, "Request"), "Method"), "OPTIONS") {
    c.AbortWithStatus(204)
    return
  }
  c.Next()
}
func loadConfig() {
  var config map[string]any = JsonParse(fs.ReadFile("./config.json")).(map[string]any)
  authKey = OSLcastString(OSLgetItem(config, "authKey"))
  useSubscriptions = OSLcastBool(OSLgetItem(config, "useSubscriptions"))
  subscriptionSizes = OSLgetItem(config, "subscriptionSizes").(map[string]any)
  quotas = OSLgetItem(config, "quotas").(map[string]any)
  if OSLnotEqual(OSLgetItem(config, "downscaleWhen"), nil) {
    downscaleWhen = OSLgetItem(config, "downscaleWhen").(map[string]any)
  }
}
func getAble(username string) map[string]any {
  username = strings.ToLower(username)
  var profile map[string]any = OSLcastObject(OSLgetItem(userData, username))
  if OSLequal(profile, nil) {
    if fs.Exists((("db/" + username) + "/user.json")) {
      var data string = OSLcastString(fs.ReadFile((("db/" + username) + "/user.json")))
      profile = JsonParse(data).(map[string]any)
    } else {
      return  map[string]any{
        "canAccess": false,
        "maxUpload": "0",
      }
    }
  }
  var subscription string = strings.ToLower(OSLcastString(OSLgetItem(profile, "subscription")))
  var quota float64 = 0
  var maybeQuota any = OSLgetItem(quotas, strings.ToLower(username))
  if OSLnotEqual(maybeQuota, nil) {
    quota = OSLcastNumber(maybeQuota)
  } else if useSubscriptions {
    quota = OSLcastNumber(OSLgetItem(subscriptionSizes, subscription))
  }
  var images []any = readUserImages(username)
  return  map[string]any{
    "canAccess": OSLcastNumber(quota) > OSLcastNumber(0),
    "storageQuota": quota,
    "hasImages": OSLcastNumber(OSLlen(images)) > OSLcastNumber(0),
  }
}
func getProfile(username string) map[string]any {
  var resp_profile map[string]any = OSLcastObject(requests.Get(("https://api.rotur.dev/profile?include_posts=0&name=" + username)))
  if (OSLgetItem(resp_profile, "success") != true) {
    return  map[string]any{
      "ok": false,
      "error": "failed to fetch profile",
    }
  }
  var profile map[string]any = OSLcastObject(JsonParse(OSLcastString(OSLgetItem(resp_profile, "body"))))
  if OSLnotEqual(OSLgetItem(profile, "error"), nil) {
    return  map[string]any{
      "ok": false,
      "error": OSLgetItem(profile, "error"),
    }
  }
  return  map[string]any{
    "ok": true,
    "profile": profile,
  }
}
func writeProfile(username string) map[string]any {
  var profileReq map[string]any = getProfile(username)
  if (OSLgetItem(profileReq, "ok") != true) {
    return profileReq
  }
  OSLsetItem(userData, strings.ToLower(username), OSLgetItem(profileReq, "profile"))
  fs.MkdirAll(("db/" + strings.ToLower(username)))
  fs.WriteFile((("db/" + strings.ToLower(username)) + "/user.json"), OSLcastString(OSLgetItem(profileReq, "profile")))
  return profileReq
}
func userDbPath(username string) string {
  return (("db/" + strings.ToLower(username)) + "/images.json")
}
func ensureUserDb(username string) {
  var dir string = ("db/" + strings.ToLower(username))
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
  var path string = userDbPath(username)
  if (fs.Exists(path) != true) {
    fs.WriteFile(path, "[]")
  }
  ensureCacheDir(username)
}
func readUserImages(username string) []any {
  ensureUserDb(username)
  var path string = userDbPath(username)
  var data string = OSLcastString(fs.ReadFile(path))
  if OSLequal(data, "") {
    return []any{}
  }
  var arr []any = OSLcastArray(JsonParse(data))
  return arr
}
func enrichImagesWithSharing(username string, images []any) []any {
  var sharesObj map[string]any = readUserShares(username)
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  for i := 1; i <= OSLlen(images); i++ {
    var img map[string]any = OSLcastObject(OSLgetItem(images, i))
    for j := 1; j <= OSLlen(shares); j++ {
      var share map[string]any = OSLcastObject(OSLgetItem(shares, j))
      if OSLequal(OSLgetItem(share, "imageId"), OSLgetItem(img, "id")) {
        OSLsetItem(img, "sharedWith", OSLgetItem(share, "sharedWith"))
        OSLsetItem(img, "isPublic", OSLgetItem(share, "isPublic"))
      }
    }
  }
  return images
}
func writeUserImages(username string, arr []any) bool {
  var path string = userDbPath(username)
  var out string = OSLcastString(arr)
  return fs.WriteFile(path, out)
}
func findImage(arr []any, id string) map[string]any {
  for i := 1; i <= OSLlen(arr); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    if OSLequal(OSLgetItem(it, "id"), id) {
      return it
    }
  }
  return map[string]any{}
}
func removeImage(arr []any, id string) []any {
  var out []any = []any{}
  for i := 1; i <= OSLlen(arr); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    if OSLnotEqual(OSLgetItem(it, "id"), id) {
      OSLappend(&(out), it)
    }
  }
  return out
}
func removeImages(arr []any, ids []any) []any {
  var out []any = []any{}
  for i := 1; i <= OSLlen(arr); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    if (OSLcontains(ids, OSLgetItem(it, "id")) != true) {
      OSLappend(&(out), it)
    }
  }
  return out
}
func userAlbumsPath(username string) string {
  return (("db/" + strings.ToLower(username)) + "/albums.json")
}
func ensureUserAlbums(username string) {
  var dir string = ("db/" + strings.ToLower(username))
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
  var path string = userAlbumsPath(username)
  if (fs.Exists(path) != true) {
    fs.WriteFile(path, "{ \"albums\": [], \"items\": {} }")
  }
}
func readUserAlbums(username string) map[string]any {
  ensureUserAlbums(username)
  var path string = userAlbumsPath(username)
  var data string = OSLcastString(fs.ReadFile(path))
  if OSLequal(data, "") {
    return  map[string]any{
      "albums": []any{},
      "items": map[string]any{},
    }
  }
  var obj map[string]any = OSLcastObject(JsonParse(data))
  if OSLequal(OSLgetItem(obj, "albums"), nil) {
    OSLsetItem(obj, "albums", []any{})
  }
  if OSLequal(OSLgetItem(obj, "items"), nil) {
    OSLsetItem(obj, "items", map[string]any{})
  }
  return obj
}
func writeUserAlbums(username string, albums map[string]any) bool {
  var path string = userAlbumsPath(username)
  var out string = OSLcastString(albums)
  return fs.WriteFile(path, out)
}
func addAlbum(username string, name string) map[string]any {
  var albums map[string]any = readUserAlbums(username)
  var list []any = OSLcastArray(OSLgetItem(albums, "albums"))
  var exists bool = false
  for i := 1; i <= OSLlen(list); i++ {
    if OSLequal(strings.ToLower(OSLcastString(OSLgetItem(list, i))), strings.ToLower(name)) {
      exists = true
    }
  }
  if (exists != true) {
    OSLappend(&(list), name)
    OSLsetItem(albums, "albums", list)
  }
  if OSLequal(OSLgetItem(OSLgetItem(albums, "items"), name), nil) {
    var items map[string]any = OSLcastObject(OSLgetItem(albums, "items"))
    OSLsetItem(items, name, []any{})
  }
  writeUserAlbums(username, albums)
  return albums
}
func removeAlbumDef(username string, name string) map[string]any {
  var albums map[string]any = readUserAlbums(username)
  var list []any = OSLcastArray(OSLgetItem(albums, "albums"))
  var out []any = []any{}
  for i := 1; i <= OSLlen(list); i++ {
    var it string = OSLcastString(OSLgetItem(list, i))
    if OSLnotEqual(strings.ToLower(it), strings.ToLower(name)) {
      OSLappend(&(out), it)
    }
  }
  OSLsetItem(albums, "albums", out)
  OSLdelete(OSLgetItem(albums, "items"), name)
  writeUserAlbums(username, albums)
  return albums
}
func addImageToAlbum(username string, name string, id string) map[string]any {
  var albums map[string]any = readUserAlbums(username)
  var items map[string]any = OSLcastObject(OSLgetItem(albums, "items"))
  if OSLequal(OSLgetItem(OSLgetItem(albums, "items"), name), nil) {
    OSLsetItem(items, name, []any{})
  }
  var ids []any = OSLcastArray(OSLgetItem(items, name))
  var exists bool = false
  for i := 1; i <= OSLlen(ids); i++ {
    if OSLequal(OSLcastString(OSLgetItem(ids, i)), id) {
      exists = true
    }
  }
  if (exists != true) {
    OSLappend(&(ids), id)
    OSLsetItem(items, name, ids)
    writeUserAlbums(username, albums)
  }
  return albums
}
func removeImageFromAlbum(username string, name string, id string) map[string]any {
  var albums map[string]any = readUserAlbums(username)
  var ids []any = OSLcastArray(OSLgetItem(OSLgetItem(albums, "items"), name))
  var out []any = []any{}
  for i := 1; i <= OSLlen(ids); i++ {
    if OSLnotEqual(OSLcastString(OSLgetItem(ids, i)), id) {
      OSLappend(&(out), OSLgetItem(ids, i))
    }
  }
  var items map[string]any = OSLcastObject(OSLgetItem(albums, "items"))
  OSLsetItem(items, name, out)
  writeUserAlbums(username, albums)
  return albums
}
func userBinPath(username string) string {
  return (("db/" + strings.ToLower(username)) + "/bin.json")
}
func ensureUserBin(username string) {
  var dir string = ("db/" + strings.ToLower(username))
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
  var path string = userBinPath(username)
  if (fs.Exists(path) != true) {
    fs.WriteFile(path, "[]")
  }
}
func ensureCacheDir(username string) {
  var dir string = (("db/" + strings.ToLower(username)) + "/cache")
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
}
func readUserBin(username string) []any {
  ensureUserBin(username)
  var path string = userBinPath(username)
  var data string = OSLcastString(fs.ReadFile(path))
  if OSLequal(data, "") {
    return []any{}
  }
  var arr []any = OSLcastArray(JsonParse(data))
  return arr
}
func writeUserBin(username string, arr []any) bool {
  var path string = userBinPath(username)
  var out string = OSLcastString(arr)
  return fs.WriteFile(path, out)
}
func calculateStorageStats(username string) map[string]any {
  var path string = (("db/" + strings.ToLower(username)) + "/blob")
  if (fs.Exists(path) != true) {
    return  map[string]any{
      "totalBytes": 0,
      "imageCount": 0,
      "largestImages": []any{},
      "binBytes": 0,
      "duplicateGroups": []any{},
    }
  }
  var files []any = OSLcastArray(fs.ReadDir(path))
  var totalBytes float64 = 0
  var imageCount int = int(0)
  var entries []any = []any{}
  var sizeGroups map[string]any = map[string]any{}
  for i := 1; i <= OSLlen(files); i++ {
    var name string = OSLcastString(OSLgetItem(files, i))
    if strings.HasSuffix(name, ".jpg") {
      var fpath string = ((path + "/") + name)
      var stat map[string]any = OSLcastObject(fs.GetStat(fpath))
      var size float64 = OSLcastNumber(OSLgetItem(stat, "size"))
      totalBytes = OSLadd(totalBytes, size)
      imageCount = OSLadd(imageCount, 1)
      var id string = OSLcastString(OSLtrim(name, 1, -5))
      var sizeStr string = OSLcastString(size)
      if OSLequal(OSLgetItem(sizeGroups, sizeStr), nil) {
        OSLsetItem(sizeGroups, sizeStr, []any{})
      }
      var sg []any = OSLcastArray(OSLgetItem(sizeGroups, sizeStr))
      OSLappend(&(sg),  map[string]any{
        "id": id,
        "path": fpath,
      })
      OSLsetItem(sizeGroups, sizeStr, sg)
      OSLappend(&(entries),  map[string]any{
        "id": id,
        "bytes": size,
        "path": fpath,
      })
    }
  }
  var hashGroups map[string]any = map[string]any{}
  var duplicateGroups []any = []any{}
  var sizeKeys []any = OSLgetKeys(sizeGroups)
  for i := 1; i <= OSLlen(sizeKeys); i++ {
    var sKey string = OSLcastString(OSLgetItem(sizeKeys, i))
    var group []any = OSLcastArray(OSLgetItem(sizeGroups, sKey))
    if OSLcastNumber(OSLlen(group)) > OSLcastNumber(1) {
      for j := 1; j <= OSLlen(group); j++ {
        var it map[string]any = OSLcastObject(OSLgetItem(group, j))
        var h string = calculateFileHash(OSLcastString(OSLgetItem(it, "path")))
        if OSLequal(OSLgetItem(hashGroups, h), nil) {
          OSLsetItem(hashGroups, h, []any{})
        }
        var hg []any = OSLcastArray(OSLgetItem(hashGroups, h))
        OSLappend(&(hg), OSLcastString(OSLgetItem(it, "id")))
        OSLsetItem(hashGroups, h, hg)
      }
    }
  }
  var hashKeys []any = OSLgetKeys(hashGroups)
  for i := 1; i <= OSLlen(hashKeys); i++ {
    var hKey string = OSLcastString(OSLgetItem(hashKeys, i))
    var group []any = OSLcastArray(OSLgetItem(hashGroups, hKey))
    if OSLcastNumber(OSLlen(group)) > OSLcastNumber(1) {
      OSLappend(&(duplicateGroups),  map[string]any{
        "hash": hKey,
        "ids": group,
      })
    }
  }
  var n int = OSLlen(entries)
  for i := 1; i <= OSLround(n); i++ {
    for j := 1; j <= int((OSLsub(n, i) - 1)); j++ {
      if OSLcastNumber(OSLcastNumber(OSLgetItem(OSLgetItem(entries, j), "bytes"))) < OSLcastNumber(OSLcastNumber(OSLgetItem(OSLgetItem(entries, OSLadd(j, 1)), "bytes"))) {
        var tmp map[string]any = OSLcastObject(OSLgetItem(entries, j))
        OSLsetItem(entries, j, OSLgetItem(entries, OSLadd(j, 1)))
        OSLsetItem(entries, OSLadd(j, 1), tmp)
      }
    }
  }
  var largest []any = []any{}
  var limit int = int(10)
  if OSLcastNumber(n) < OSLcastNumber(limit) {
    limit = n
  }
  for i := 1; i <= OSLround(limit); i++ {
    OSLappend(&(largest),  map[string]any{
      "id": OSLgetItem(OSLgetItem(entries, i), "id"),
      "bytes": OSLgetItem(OSLgetItem(entries, i), "bytes"),
    })
  }
  var binPath string = (("db/" + strings.ToLower(username)) + "/bin")
  var binBytes float64 = 0
  if fs.Exists(binPath) {
    var binFiles []any = OSLcastArray(fs.ReadDir(binPath))
    for i := 1; i <= OSLlen(binFiles); i++ {
      var bname string = OSLcastString(OSLgetItem(binFiles, i))
      if strings.HasSuffix(bname, ".jpg") {
        var bfpath string = ((binPath + "/") + bname)
        var bstat map[string]any = OSLcastObject(fs.GetStat(bfpath))
        binBytes = OSLadd(binBytes, OSLcastNumber(OSLgetItem(bstat, "size")))
      }
    }
  }
  var fileSizes map[string]any = map[string]any{}
  for i := 1; i <= OSLlen(entries); i++ {
    OSLsetItem(fileSizes, OSLcastString(OSLgetItem(OSLgetItem(entries, i), "id")), OSLgetItem(OSLgetItem(entries, i), "bytes"))
  }
  return  map[string]any{
    "totalBytes": totalBytes,
    "imageCount": imageCount,
    "largestImages": largest,
    "binBytes": binBytes,
    "duplicateGroups": duplicateGroups,
    "fileSizes": fileSizes,
  }
}
func userSharesPath(username string) string {
  return (("db/" + strings.ToLower(username)) + "/shares.json")
}
func ensureUserShares(username string) {
  var dir string = ("db/" + strings.ToLower(username))
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
  var path string = userSharesPath(username)
  if (fs.Exists(path) != true) {
    fs.WriteFile(path, "{ \"shares\": [] }")
  }
}
func readUserShares(username string) map[string]any {
  ensureUserShares(username)
  var path string = userSharesPath(username)
  var data string = OSLcastString(fs.ReadFile(path))
  if OSLequal(data, "") {
    return  map[string]any{
      "shares": []any{},
    }
  }
  var obj map[string]any = OSLcastObject(JsonParse(data))
  if OSLequal(OSLgetItem(obj, "shares"), nil) {
    OSLsetItem(obj, "shares", []any{})
  }
  return obj
}
func writeUserShares(username string, sharesObj map[string]any) bool {
  var path string = userSharesPath(username)
  return fs.WriteFile(path, OSLcastString(sharesObj))
}
func addShare(owner string, imageId string, targetUsername string) map[string]any {
  var sharesObj map[string]any = readUserShares(owner)
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  var found bool = false
  for i := 1; i <= OSLlen(shares); i++ {
    var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
    if OSLequal(OSLcastString(OSLgetItem(share, "imageId")), imageId) {
      var sharedWith []any = OSLcastArray(OSLgetItem(share, "sharedWith"))
      var exists bool = false
      for j := 1; j <= OSLlen(sharedWith); j++ {
        if OSLequal(strings.ToLower(OSLcastString(OSLgetItem(sharedWith, j))), strings.ToLower(targetUsername)) {
          exists = true
        }
      }
      if (exists != true) {
        OSLappend(&(sharedWith), strings.ToLower(targetUsername))
        OSLsetItem(share, "sharedWith", sharedWith)
      }
      found = true
    }
  }
  if (found != true) {
    OSLappend(&(shares),  map[string]any{
      "imageId": imageId,
      "sharedWith": []any{
        strings.ToLower(targetUsername),
      },
    })
  }
  OSLsetItem(sharesObj, "shares", shares)
  writeUserShares(owner, sharesObj)
  return sharesObj
}
func removeShare(owner string, imageId string, targetUsername string) map[string]any {
  var sharesObj map[string]any = readUserShares(owner)
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  var newShares []any = []any{}
  for i := 1; i <= OSLlen(shares); i++ {
    var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
    if OSLequal(OSLcastString(OSLgetItem(share, "imageId")), imageId) {
      var sharedWith []any = OSLcastArray(OSLgetItem(share, "sharedWith"))
      var newSharedWith []any = []any{}
      for j := 1; j <= OSLlen(sharedWith); j++ {
        if OSLnotEqual(strings.ToLower(OSLcastString(OSLgetItem(sharedWith, j))), strings.ToLower(targetUsername)) {
          OSLappend(&(newSharedWith), OSLgetItem(sharedWith, j))
        }
      }
      if OSLcastNumber(OSLlen(newSharedWith)) > OSLcastNumber(0) {
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
func setPublicShare(owner string, imageId string, isPublic bool) map[string]any {
  var sharesObj map[string]any = readUserShares(owner)
  var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
  var found bool = false
  for i := 1; i <= OSLlen(shares); i++ {
    var share map[string]any = OSLcastObject(OSLgetItem(shares, i))
    if OSLequal(OSLcastString(OSLgetItem(share, "imageId")), imageId) {
      OSLsetItem(share, "isPublic", isPublic)
      found = true
    }
  }
  if (found != true) && OSLequal(isPublic, true) {
    OSLappend(&(shares),  map[string]any{
      "imageId": imageId,
      "sharedWith": []any{},
      "isPublic": true,
    })
  }
  OSLsetItem(sharesObj, "shares", shares)
  writeUserShares(owner, sharesObj)
  return sharesObj
}
func getSharedWithMe(username string) []any {
  var results []any = []any{}
  var dbPath string = "db"
  if (fs.Exists(dbPath) != true) {
    return results
  }
  var dirs []any = OSLcastArray(fs.ReadDir(dbPath))
  for i := 1; i <= OSLlen(dirs); i++ {
    var ownerDir string = OSLcastString(OSLgetItem(dirs, i))
    var sharesPath string = (((dbPath + "/") + ownerDir) + "/shares.json")
    if fs.Exists(sharesPath) {
      var data string = OSLcastString(fs.ReadFile(sharesPath))
      if OSLnotEqual(data, "") {
        var sharesObj map[string]any = OSLcastObject(JsonParse(data))
        var shares []any = OSLcastArray(OSLgetItem(sharesObj, "shares"))
        if OSLnotEqual(shares, nil) {
          for j := 1; j <= OSLlen(shares); j++ {
            var share map[string]any = OSLcastObject(OSLgetItem(shares, j))
            var sharedWith []any = OSLcastArray(OSLgetItem(share, "sharedWith"))
            for k := 1; k <= OSLlen(sharedWith); k++ {
              if OSLequal(strings.ToLower(OSLcastString(OSLgetItem(sharedWith, k))), strings.ToLower(username)) {
                OSLappend(&(results),  map[string]any{
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
func getCachePath(username string, id string) string {
  return (((("db/" + strings.ToLower(username)) + "/cache/") + id) + "_preview.jpg")
}
func servePreview(c *gin.Context, path string, id string, username string) bool {
  if (fs.Exists(path) != true) || fs.IsDir(path) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return false
  }
  var cachePath string = getCachePath(username, id)
  if fs.Exists(cachePath) {
    var cacheInfo map[string]any = OSLcastObject(fs.GetStat(cachePath))
    var origInfo map[string]any = OSLcastObject(fs.GetStat(path))
    if OSLcastNumber(OSLcastNumber(OSLgetItem(cacheInfo, "modTime"))) >= OSLcastNumber(OSLcastNumber(OSLgetItem(origInfo, "modTime"))) {
      var etag string = ("preview-" + OSLcastString(OSLgetItem(cacheInfo, "modTime")))
      var ifNoneMatch string = OSLcastString(c.GetHeader("If-None-Match"))
      if OSLequal(ifNoneMatch, etag) {
        c.Status(304)
        return true
      }
      c.Header("ETag", etag)
      c.Header("Cache-Control", "private, max-age=31536000, immutable")
      c.File(cachePath)
      return true
    }
  }
  var fileInfo map[string]any = OSLcastObject(fs.GetStat(path))
  var fileSize float64 = OSLcastNumber(OSLgetItem(fileInfo, "size"))
  if OSLequal(fileSize, 0) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return false
  }
  if OSLcastNumber(fileSize) > OSLcastNumber(1e+07) {
    c.JSON(413,  map[string]any{
      "ok": false,
      "error": "file too large",
    })
    return false
  }
  var im = img.Open(path)
  if OSLequal(im, nil) {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid image",
    })
    return false
  }
  defer im.Close()
  var w float64 = OSLcastNumber(im.Width())
  var h float64 = OSLcastNumber(im.Height())
  var maxw float64 = 200
  var maxh float64 = 200
  if OSLcastNumber(w) <= OSLcastNumber(maxw) && OSLcastNumber(h) <= OSLcastNumber(maxh) {
    c.Header("Cache-Control", "private, max-age=31536000, immutable")
    c.File(path)
    return true
  }
  var ratio float64 = OSLmin(OSLdivide(maxw, w), OSLdivide(maxh, h))
  var rw int = OSLround(OSLmultiply(w, ratio))
  var rh int = OSLround(OSLmultiply(h, ratio))
  var resized = img.ResizeFast(im, rw, rh)
  if OSLequal(resized, nil) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "resize failed",
    })
    return false
  }
  defer resized.Close()
  var cached bool = bool(img.SaveJPEG(resized, cachePath, 80))
  if cached {
    c.Header("Cache-Control", "private, max-age=31536000, immutable")
    c.File(cachePath)
  } else {
    var out []byte = img.EncodeJPEGBytes(resized, 80)
    if OSLequal(OSLlen(out), 0) {
      c.JSON(500,  map[string]any{
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















var sessions map[string]any = map[string]any{}
var userData map[string]any = map[string]any{}
var useSubscriptions bool = false
var subscriptionSizes map[string]any = map[string]any{}
var quotas map[string]any = map[string]any{}
var downscaleWhen map[string]any = map[string]any{}
var authKey string = ""
func up(c *gin.Context) {
  c.String(200, "ok")
}
func main() {
  if fs.Exists("db/sessions.json") {
    var data string = OSLcastString(fs.ReadFile("db/sessions.json"))
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
  image.POST("/upload", requireSession, handleUpload)
  api.POST("/image", requireSession, handleUpload)
  image.DELETE("/:id", requireSession, handleDeleteImage)
  image.GET("/:id", requireSession, handleId)
  image.POST("/:id/rotate", requireSession, handleImageRotate)
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
  r.Run(":5607")
}

