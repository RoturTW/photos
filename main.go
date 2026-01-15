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
	"github.com/gin-gonic/gin"
	godotenv "github.com/joho/godotenv"
	"github.com/rwcarlsen/goexif/exif"
	"path/filepath"
	OSL_bytes "bytes"
	OSL_draw "golang.org/x/image/draw"
	OSL_image "image"
	"image/png"
	"image/jpeg"
	"sync"
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
	case map[string]any:
		return JsonStringify(s)
	case map[string]string:
		return JsonStringify(s)
	case map[string]int:
		return JsonStringify(s)
	case map[string]float64:
		return JsonStringify(s)
	case map[string]bool:
		return JsonStringify(s)
	case io.Reader:
		data, err := io.ReadAll(s)
		if err != nil {
			panic("OSLcastString: failed to read io.Reader:" + err.Error())
			return ""
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

func OSLrandom(low any, high any) int {
	highInt := OSLcastInt(high)
	lowInt := OSLcastInt(low)
	return OSLcastInt(rand.Intn(int(highInt-lowInt+1))) + lowInt
}

func OSLnullishCoaless(a any, b any) any {
	if a == nil {
		return b
	}
	return a
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

// Helper function to convert bool to float64
func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// Helper function to convert bool to int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

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

func OSLjoin(a any, b any) string {
	a = OSLcastString(a)
	b = OSLcastString(b)
	return OSLcastString(a) + OSLcastString(b)
}

// OSLmultiply handles the * operation: multiplies numbers, repeats strings
func OSLmultiply(a any, b any) float64 {
	return OSLcastNumber(a) * OSLcastNumber(b)
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

func (h *HTTP) Get(url string, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodGet, url, m)
}

func (h *HTTP) Post(url string, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPost, url, data)
}

func (h *HTTP) Put(url string, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPut, url, data)
}

func (h *HTTP) Patch(url string, data map[string]any) map[string]any {
	return h.doRequest(http.MethodPatch, url, data)
}

func (h *HTTP) Delete(url string, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodDelete, url, m)
}

func (h *HTTP) Options(url string, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	return h.doRequest(http.MethodOptions, url, m)
}

func (h *HTTP) Head(url string, data ...map[string]any) map[string]any {
	var m map[string]any
	if len(data) > 0 {
		m = data[0]
	}
	out := map[string]any{"success": false}
	headers, _ := extractHeadersAndBody(m)
	req, err := http.NewRequest(http.MethodHead, url, nil)
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
// description: Dynamic image utilities for OSL
// author: Mist
// requires: bytes as OSL_bytes, golang.org/x/image/draw as OSL_draw, image as OSL_image, image/png, image/jpeg, sync

type IMG struct{}

var (
	OSL_img_Store = map[string]OSL_image.Image{}
	OSL_img_Mu    sync.Mutex
)

func OSL_img_store(im OSL_image.Image) string {
	id := fmt.Sprintf("img_%d", time.Now().UnixNano())
	OSL_img_Mu.Lock()
	OSL_img_Store[id] = im
	OSL_img_Mu.Unlock()
	return id
}

func OSL_img_get(id string) OSL_image.Image {
	OSL_img_Mu.Lock()
	im := OSL_img_Store[id]
	OSL_img_Mu.Unlock()

	return im
}

func (IMG) GetImage(id string) OSL_image.Image {
	return OSL_img_get(id)
}

func (IMG) UseImage(imgage OSL_image.Image) string {
	return OSL_img_store(imgage)
}

func (IMG) DecodeBytes(data []byte) string {
	im, _, err := OSL_image.Decode(OSL_bytes.NewReader(data))
	if err != nil {
		return ""
	}

	return OSL_img_store(im)
}

func (IMG) DecodeFile(path any) string {
	b, err := os.ReadFile(OSLcastString(path))
	if err != nil {
		return ""
	}
	return img.DecodeBytes(b)
}

func (IMG) EncodePNG(id any) []byte {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return []byte{}
	}

	var buf OSL_bytes.Buffer
	_ = png.Encode(&buf, im)
	return buf.Bytes()
}

func (IMG) EncodeJPEG(id any, quality any) []byte {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return []byte{}
	}

	v := int(OSLcastNumber(quality))

	var buf OSL_bytes.Buffer
	_ = jpeg.Encode(&buf, im, &jpeg.Options{Quality: v})
	return buf.Bytes()
}

func (IMG) Size(id any) map[string]any {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return map[string]any{"w": 0, "h": 0}
	}

	b := im.Bounds()
	return map[string]any{
		"w": b.Dx(),
		"h": b.Dy(),
	}
}

func (IMG) Bounds(id any) map[string]any {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return map[string]any{}
	}

	b := im.Bounds()
	return map[string]any{
		"minX": b.Min.X,
		"minY": b.Min.Y,
		"maxX": b.Max.X,
		"maxY": b.Max.Y,
	}
}

func (IMG) SavePNG(id any, path any) bool {
	data := img.EncodePNG(OSLcastString(id))
	return os.WriteFile(OSLcastString(path), data, 0644) == nil
}

func (IMG) SaveJPEG(id any, path any, quality any) bool {
	data := img.EncodeJPEG(OSLcastString(id), OSLcastNumber(quality))
	return os.WriteFile(OSLcastString(path), data, 0644) == nil
}

func (IMG) Resize(id any, width any, height any) string {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return ""
	}

	w := int(OSLcastNumber(width))
	h := int(OSLcastNumber(height))

	if w <= 0 || h <= 0 {
		return ""
	}

	dst := OSL_image.NewRGBA(OSL_image.Rect(0, 0, w, h))

	OSL_draw.CatmullRom.Scale(dst, dst.Bounds(), im, im.Bounds(), OSL_draw.Over, nil)

	return OSL_img_store(dst)
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
  var resp map[string]any = OSLcastObject(requests.Get("https://api.rotur.dev/validate?key=rotur-gate&v=" + token))
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
  var username string = OSLcastString(OSLgetItem(strings.Split(token, ","), 1))
  OSLsetItem(sessions, sessionId, username)
  var resp_profile map[string]any = OSLcastObject(requests.Get("https://api.rotur.dev/profile?name=" + username))
  if (OSLgetItem(resp_profile, "success") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "failed to fetch profile",
    })
    return
  }
  var profile map[string]any = OSLcastObject(JsonParse(OSLcastString(OSLgetItem(resp_profile, "body"))))
  if OSLnotEqual(OSLgetItem(profile, "error"), nil) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": OSLgetItem(profile, "error"),
    })
    return
  }
  OSLsetItem(userData, strings.ToLower(username), profile)
  c.SetCookie("session_id", sessionId, 3600, "/", "", false, true)
  c.JSON(200,  map[string]any{
    "ok": true,
    "token": token,
  })
}
func handleAble(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  c.JSON(200, able)
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
  var n int = OSLcastInt(OSLlen(arr))
  var nowMs float64 = OSLcastNumber(time.Now().UnixMilli())
  var ninety float64 = OSLcastNumber(OSLmultiply(OSLmultiply(OSLmultiply(OSLmultiply(90, 24), 60), 60), 1000))
  var out []any = []any{}
  for i := 1; i <= n; i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    var ts float64 = OSLcastNumber(OSLgetItem(it, "timestamp"))
    if OSLcastNumber((nowMs - ts)) <= OSLcastNumber(ninety) {
      OSLappend(&(out), it)
    }
  }
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
  var year int = OSLcastInt(OSLcastNumber(c.Param("year")))
  var arr []any = OSLcastArray(readUserImages(username))
  var n int = OSLcastInt(OSLlen(arr))
  var out []any = []any{}
  for i := 1; i <= n; i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    var ts int = OSLcastInt(OSLcastNumber(OSLgetItem(it, "timestamp")))
    var t = time.UnixMilli(int64(ts))
    if OSLequal(t.Year(), year) {
      OSLappend(&(out), it)
    }
  }
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
  var year int = OSLcastInt(OSLcastNumber(c.Param("year")))
  var month int = OSLcastInt(OSLcastNumber(c.Param("month")))
  var arr []any = OSLcastArray(readUserImages(username))
  var n int = OSLcastInt(OSLlen(arr))
  var out []any = []any{}
  for i := 1; i <= n; i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    var ts int = OSLcastInt(OSLcastNumber(OSLgetItem(it, "timestamp")))
    var t = time.UnixMilli(int64(ts))
    if OSLequal(t.Year(), year) && OSLequal(int(t.Month()), month) {
      OSLappend(&(out), it)
    }
  }
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
  var tmpBase string = OSLcastString(OSLjoin(OSLjoin("db/", strings.ToLower(username)), "/blob"))
  if (fs.Exists(tmpBase) != true) {
    fs.MkdirAll(tmpBase)
  }
  var tmpPath string = OSLcastString(OSLjoin(OSLjoin(tmpBase, "/tmp-"), randomString(16)))
  var body string = OSLcastString(OSLgetItem(OSLgetItem(c, "Request"), "Body"))
  if OSLequal(body, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "empty body",
    })
    return
  }
  var wroteTmp bool = fs.WriteFile(tmpPath, body)
  if (wroteTmp != true) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "failed to write upload",
    })
    return
  }
  var imgId string = OSLcastString(img.DecodeFile(tmpPath))
  if OSLequal(imgId, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid image",
    })
    fs.Remove(tmpPath)
    return
  }
  var size map[string]any = OSLcastObject(img.Size(imgId))
  var w int = OSLcastInt(OSLround(OSLgetItem(size, "w")))
  var h int = OSLcastInt(OSLround(OSLgetItem(size, "h")))
  var id string = OSLcastString(randomString(24))
  var path string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(tmpBase, "/"), id), ".jpg"))
  var rid string = OSLcastString(img.Resize(imgId, w, h))
  var wrote bool = img.SaveJPEG(rid, path, 90)
  if (wrote != true) {
    c.JSON(500,  map[string]any{
      "ok": false,
      "error": "failed to save",
    })
    fs.Remove(tmpPath)
    return
  }
  var tsms float64 = OSLcastNumber(time.Now().UnixMilli())
  f, err := os.Open(tmpPath)
  if OSLequal(err, nil) {
    x, err := exif.Decode(f)
    if OSLequal(err, nil) {
      t0, err := x.DateTime()
      if OSLequal(err, nil) && (t0.IsZero() != true) {
        tsms = OSLcastNumber(t0.UnixMilli())
      } else {
        tag, err := x.Get(exif.DateTimeOriginal)
        if OSLequal(err, nil) && OSLnotEqual(tag, nil) {
          s, err := tag.StringVal()
          if OSLequal(err, nil) {
            t1, err := time.Parse("2006:01:02 15:04:05", s)
            if OSLequal(err, nil) && (t1.IsZero() != true) {
              tsms = OSLcastNumber(t1.UnixMilli())
            }
          }
        }
      }
    }
  }
  fs.Remove(tmpPath)
  var arr []any = OSLcastArray(readUserImages(username))
  var entry map[string]any = OSLcastObject( map[string]any{
    "id": id,
    "width": w,
    "height": h,
    "timestamp": tsms,
  })
  OSLappend(&(arr), entry)
  writeUserImages(username, arr)
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
  var json bool = bool(false)
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
  var path string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(OSLjoin("db/", strings.ToLower(username)), "/blob/"), id), ".jpg"))
  var data []byte = fs.ReadFileBytes(path)
  if OSLequal(OSLlen(data), 0) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  c.Data(200, "image/jpeg", data)
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
  var q string = OSLcastString(strings.ToLower(OSLcastString(c.Query("q"))))
  var arr []any = OSLcastArray(readUserImages(username))
  if OSLequal(q, "") {
    c.JSON(200, arr)
    return
  }
  var out []any = []any{}
  var n int = OSLcastInt(OSLlen(arr))
  for i := 1; i <= n; i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    if OSLcontains(strings.ToLower(OSLcastString(OSLgetItem(it, "id"))), q) {
      OSLappend(&(out), it)
    }
  }
  c.JSON(200, out)
}
func handleShareMine(c *gin.Context) {
  c.JSON(200, []any{})
}
func handleShareOthers(c *gin.Context) {
  c.JSON(200, []any{})
}
func handleSharePatch(c *gin.Context) {
  c.JSON(501,  map[string]any{
    "ok": false,
    "error": "not implemented",
  })
}
func handlePreview(c *gin.Context) {
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
  var path string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(OSLjoin("db/", strings.ToLower(username)), "/blob/"), id), ".jpg"))
  var data []byte = fs.ReadFileBytes(path)
  if OSLequal(OSLlen(data), 0) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  c.Header("Cache-Control", "public, max-age=31536000, immutable")
  var imgId string = OSLcastString(img.DecodeBytes(data))
  if OSLequal(imgId, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid image",
    })
    return
  }
  var s map[string]any = OSLcastObject(img.Size(imgId))
  var w float64 = OSLcastNumber(OSLgetItem(s, "w"))
  var h float64 = OSLcastNumber(OSLgetItem(s, "h"))
  var maxw float64 = OSLcastNumber(800)
  var maxh float64 = OSLcastNumber(800)
  var rw float64 = OSLcastNumber(w)
  var rh float64 = OSLcastNumber(h)
  if OSLcastNumber(w) > OSLcastNumber(maxw) {
    rw = maxw
    rh = (OSLmultiply(h, maxw) / w)
  }
  if OSLcastNumber(rh) > OSLcastNumber(maxh) {
    rh = maxh
    rw = (OSLmultiply(rw, maxh) / rh)
  }
  var rid string = OSLcastString(img.Resize(imgId, OSLround(rw), OSLround(rh)))
  var out []byte = img.EncodeJPEG(rid, 80)
  c.Data(200, "image/jpeg", out)
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
  var base string = OSLcastString(OSLjoin("db/", strings.ToLower(username)))
  var path string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(base, "/blob/"), id), ".jpg"))
  var binBase string = OSLcastString(OSLjoin(base, "/bin"))
  if (fs.Exists(binBase) != true) {
    fs.MkdirAll(binBase)
  }
  var data []byte = fs.ReadFileBytes(path)
  if OSLcastNumber(OSLlen(data)) > OSLcastNumber(0) {
    var imgId string = OSLcastString(img.DecodeBytes(data))
    if OSLnotEqual(imgId, "") {
      var s map[string]any = OSLcastObject(img.Size(imgId))
      var w int = OSLcastInt(OSLround(OSLgetItem(s, "w")))
      var h int = OSLcastInt(OSLround(OSLgetItem(s, "h")))
      var rid string = OSLcastString(img.Resize(imgId, w, h))
      var binPath string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(binBase, "/"), id), ".jpg"))
      var wrote bool = img.SaveJPEG(rid, binPath, 90)
      if wrote {
        var binArr []any = OSLcastArray(readUserBin(username))
        var tsms float64 = OSLcastNumber(time.Now().UnixMilli())
        var arr []any = OSLcastArray(readUserImages(username))
        var it map[string]any = OSLcastObject(findImage(arr, id))
        var origTs float64 = OSLcastNumber(OSLgetItem(it, "timestamp"))
        if OSLequal(origTs, 0) {
          origTs = tsms
        }
        var entry map[string]any = OSLcastObject( map[string]any{
          "id": id,
          "width": w,
          "height": h,
          "timestamp": origTs,
          "deletedAt": tsms,
        })
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
  var base string = OSLcastString(OSLjoin("db/", strings.ToLower(username)))
  var binPath string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(base, "/bin/"), id), ".jpg"))
  var blobPath string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(base, "/blob/"), id), ".jpg"))
  var data []byte = fs.ReadFileBytes(binPath)
  if OSLequal(OSLlen(data), 0) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  var imgId string = OSLcastString(img.DecodeBytes(data))
  if OSLequal(imgId, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid image",
    })
    return
  }
  var s map[string]any = OSLcastObject(img.Size(imgId))
  var w int = OSLcastInt(OSLround(OSLgetItem(s, "w")))
  var h int = OSLcastInt(OSLround(OSLgetItem(s, "h")))
  var rid string = OSLcastString(img.Resize(imgId, w, h))
  var wrote bool = img.SaveJPEG(rid, blobPath, 90)
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
  var entry map[string]any = OSLcastObject( map[string]any{
    "id": id,
    "width": w,
    "height": h,
    "timestamp": ts,
  })
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
  var base string = OSLcastString(OSLjoin("db/", strings.ToLower(username)))
  var binPath string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(base, "/bin/"), id), ".jpg"))
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
func handleAlbums(c *gin.Context) {
  var username string = OSLcastString(c.MustGet("username"))
  var able map[string]any = OSLcastObject(getAble(username))
  if (OSLgetItem(able, "canAccess") != true) {
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
  if (OSLgetItem(able, "canAccess") != true) {
    c.JSON(401,  map[string]any{
      "ok": false,
      "error": "not authorized",
    })
    return
  }
  var id string = OSLcastString(c.Param("id"))
  var path string = OSLcastString(OSLjoin(OSLjoin(OSLjoin(OSLjoin("db/", strings.ToLower(username)), "/bin/"), id), ".jpg"))
  var data []byte = fs.ReadFileBytes(path)
  if OSLequal(OSLlen(data), 0) {
    c.JSON(404,  map[string]any{
      "ok": false,
      "error": "not found",
    })
    return
  }
  c.Header("Cache-Control", "public, max-age=31536000, immutable")
  var imgId string = OSLcastString(img.DecodeBytes(data))
  if OSLequal(imgId, "") {
    c.JSON(400,  map[string]any{
      "ok": false,
      "error": "invalid image",
    })
    return
  }
  var s map[string]any = OSLcastObject(img.Size(imgId))
  var w float64 = OSLcastNumber(OSLgetItem(s, "w"))
  var h float64 = OSLcastNumber(OSLgetItem(s, "h"))
  var maxw float64 = OSLcastNumber(800)
  var maxh float64 = OSLcastNumber(800)
  var rw float64 = OSLcastNumber(w)
  var rh float64 = OSLcastNumber(h)
  if OSLcastNumber(w) > OSLcastNumber(maxw) {
    rw = maxw
    rh = (OSLmultiply(h, maxw) / w)
  }
  if OSLcastNumber(rh) > OSLcastNumber(maxh) {
    rh = maxh
    rw = (OSLmultiply(rw, maxh) / rh)
  }
  var rid string = OSLcastString(img.Resize(imgId, OSLround(rw), OSLround(rh)))
  var out []byte = img.EncodeJPEG(rid, 80)
  c.Data(200, "image/jpeg", out)
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
  c.Set("username", OSLgetItem(sessions, sessionId))
  c.Next()
}

func homePage(c *gin.Context) {
  var username = OSLcastString(c.MustGet("username"))
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
  var chars string = OSLcastString("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
  var result string = OSLcastString("")
  for i := 1; i <= length; i++ {
    result = OSLjoin(result, OSLcastString(OSLgetItem(chars, rand.Intn(OSLlen(chars)) + 1)))
  }
  return result
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
func getAble(username string) map[string]any {
  username = strings.ToLower(username)
  var profile map[string]any = OSLcastObject(OSLgetItem(userData, username))
  if OSLequal(profile, nil) {
    return  map[string]any{
      "canAccess": false,
      "maxUpload": "0",
    }
  }
  var subscription string = OSLcastString(strings.ToLower(OSLcastString(OSLgetItem(profile, "subscription"))))
  return  map[string]any{
    "canAccess": OSLequal(subscription, "drive") || OSLequal(subscription, "pro") || OSLequal(subscription, "max"),
    "maxUpload": "15000000",
  }
}
func userDbPath(username string) string {
  return OSLjoin(OSLjoin("db/", strings.ToLower(username)), "/images.json")
}
func ensureUserDb(username string) {
  var dir string = OSLcastString(OSLjoin("db/", strings.ToLower(username)))
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
  var path string = OSLcastString(userDbPath(username))
  if (fs.Exists(path) != true) {
    fs.WriteFile(path, "[]")
  }
}
func readUserImages(username string) []any {
  ensureUserDb(username)
  var path string = OSLcastString(userDbPath(username))
  var data string = OSLcastString(fs.ReadFile(path))
  if OSLequal(data, "") {
    return []any{}
  }
  var arr []any = OSLcastArray(JsonParse(data))
  return arr
}
func writeUserImages(username string, arr []any) bool {
  var path string = OSLcastString(userDbPath(username))
  var out string = OSLcastString(arr)
  return fs.WriteFile(path, out)
}
func findImage(arr []any, id string) map[string]any {
  for i := 1; i <= OSLlen(arr); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    if OSLequal(OSLcastString(OSLgetItem(it, "id")), id) {
      return it
    }
  }
  return map[string]any{}
}
func removeImage(arr []any, id string) []any {
  var out []any = []any{}
  for i := 1; i <= OSLlen(arr); i++ {
    var it map[string]any = OSLcastObject(OSLgetItem(arr, i))
    if OSLnotEqual(OSLcastString(OSLgetItem(it, "id")), id) {
      OSLappend(&(out), it)
    }
  }
  return out
}
func userAlbumsPath(username string) string {
  return OSLjoin(OSLjoin("db/", strings.ToLower(username)), "/albums.json")
}
func ensureUserAlbums(username string) {
  var dir string = OSLcastString(OSLjoin("db/", strings.ToLower(username)))
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
  var path string = OSLcastString(userAlbumsPath(username))
  if (fs.Exists(path) != true) {
    fs.WriteFile(path, "{ \"albums\": [], \"items\": {} }")
  }
}
func readUserAlbums(username string) map[string]any {
  ensureUserAlbums(username)
  var path string = OSLcastString(userAlbumsPath(username))
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
  var path string = OSLcastString(userAlbumsPath(username))
  var out string = OSLcastString(albums)
  return fs.WriteFile(path, out)
}
func addAlbum(username string, name string) map[string]any {
  var albums map[string]any = OSLcastObject(readUserAlbums(username))
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
  if (OSLgetItem(OSLgetItem(albums, "items"), name) != true) {
    OSLsetItem(albums, name, []any{})
  }
  writeUserAlbums(username, albums)
  return albums
}
func removeAlbumDef(username string, name string) map[string]any {
  var albums map[string]any = OSLcastObject(readUserAlbums(username))
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
  var albums map[string]any = OSLcastObject(readUserAlbums(username))
  if (OSLgetItem(OSLgetItem(albums, "items"), name) != true) {
    OSLsetItem(albums, name, []any{})
  }
  var ids []any = OSLcastArray(OSLgetItem(OSLgetItem(albums, "items"), name))
  var exists bool = false
  for i := 1; i <= OSLlen(ids); i++ {
    if OSLequal(OSLcastString(OSLgetItem(ids, i)), id) {
      exists = true
    }
  }
  if (exists != true) {
    OSLappend(&(ids), id)
    OSLsetItem(albums, name, ids)
    writeUserAlbums(username, albums)
  }
  return albums
}
func removeImageFromAlbum(username string, name string, id string) map[string]any {
  var albums map[string]any = OSLcastObject(readUserAlbums(username))
  var ids []any = OSLcastArray(OSLgetItem(OSLgetItem(albums, "items"), name))
  var out []any = []any{}
  for i := 1; i <= OSLlen(ids); i++ {
    if OSLnotEqual(OSLcastString(OSLgetItem(ids, i)), id) {
      OSLappend(&(out), OSLgetItem(ids, i))
    }
  }
  OSLsetItem(albums, name, out)
  writeUserAlbums(username, albums)
  return albums
}
func userBinPath(username string) string {
  return OSLjoin(OSLjoin("db/", strings.ToLower(username)), "/bin.json")
}
func ensureUserBin(username string) {
  var dir string = OSLcastString(OSLjoin("db/", strings.ToLower(username)))
  if (fs.Exists(dir) != true) {
    fs.MkdirAll(dir)
  }
  var path string = OSLcastString(userBinPath(username))
  if (fs.Exists(path) != true) {
    fs.WriteFile(path, "[]")
  }
}
func readUserBin(username string) []any {
  ensureUserBin(username)
  var path string = OSLcastString(userBinPath(username))
  var data string = OSLcastString(fs.ReadFile(path))
  if OSLequal(data, "") {
    return []any{}
  }
  var arr []any = OSLcastArray(JsonParse(data))
  return arr
}
func writeUserBin(username string, arr []any) bool {
  var path string = OSLcastString(userBinPath(username))
  var out string = OSLcastString(arr)
  return fs.WriteFile(path, out)
}













var sessions map[string]any = map[string]any{}
var userData map[string]any = map[string]any{}
var authKey string = OSLcastString("")
func up(c *gin.Context) {
  c.String(200, "ok")
}
func main() {
  var home string = OSLcastString(os.Getenv("HOME"))
  godotenv.Overload(OSLjoin(home, "/Documents/.env"))
  godotenv.Load()
  var r = gin.Default()
  r.Use(noCORS)
  r.LoadHTMLGlob("templates/*")
  r.Static("/static", "./static")
  r.GET("/", requireSession, homePage)
  r.GET("/auth", authPage)
  var api = r.Group("/api")
  api.GET("/auth", handleAuth)
  api.GET("/able", requireSession, handleAble)
  api.GET("/search", requireSession, handleSearch)
  api.GET("/share/mine", requireSession, handleShareMine)
  api.GET("/share/others", requireSession, handleShareOthers)
  api.PATCH("/share/:id", requireSession, handleSharePatch)
  var images = api.Group("/images")
  images.GET("/recent", requireSession, handleRecentImages)
  images.GET("/:year", requireSession, handleYearImages)
  images.GET("/:year/:month", requireSession, handleMonthImages)
  var image = api.Group("/image")
  image.POST("/upload", requireSession, handleUpload)
  api.POST("/image", requireSession, handleUpload)
  image.DELETE("/:id", requireSession, handleDeleteImage)
  image.GET("/:id", requireSession, handleId)
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
  bin.GET("/preview/:id", requireSession, handleBinPreview)
  r.Run(":5606")
}

