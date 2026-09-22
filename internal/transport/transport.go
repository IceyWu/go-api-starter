// Package transport provides the application's Chi-compatible HTTP transport
// helpers, request context, response writers, and middleware adapters.
package transport

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type H map[string]interface{}

type HandlerFunc func(*Context)

type IRouter interface {
	GET(string, ...HandlerFunc)
	POST(string, ...HandlerFunc)
	PUT(string, ...HandlerFunc)
	DELETE(string, ...HandlerFunc)
	OPTIONS(string, ...HandlerFunc)
}

type Accounts map[string]string

const abortIndex = int8(63)

type Error struct{ Err error }

type Errors []*Error

func (e Errors) Last() *Error {
	if len(e) == 0 {
		return nil
	}
	return e[len(e)-1]
}

type ResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *ResponseWriter) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

func (w *ResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *ResponseWriter) Size() int { return w.size }

func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return h.Hijack()
}

func (w *ResponseWriter) Flush() {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

type Context struct {
	Request *http.Request
	Writer  *ResponseWriter

	index    int8
	handlers []HandlerFunc
	keys     map[string]interface{}
	Errors   Errors
	fullPath string
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) && c.index < abortIndex {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Abort() { c.index = abortIndex }

func (c *Context) IsAborted() bool { return c.index >= abortIndex }

func (c *Context) Error(err error) *Error {
	item := &Error{Err: err}
	c.Errors = append(c.Errors, item)
	return item
}

func (c *Context) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	value, ok := c.keys[key]
	return value, ok
}

func (c *Context) Param(name string) string { return chi.URLParam(c.Request, name) }

func (c *Context) Query(name string) string { return c.Request.URL.Query().Get(name) }

func (c *Context) GetHeader(name string) string { return c.Request.Header.Get(name) }

func (c *Context) Header(name, value string) { c.Writer.Header().Set(name, value) }

func (c *Context) ClientIP() string {
	if forwarded := c.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil {
		return host
	}
	return c.Request.RemoteAddr
}

func (c *Context) FullPath() string {
	if c.fullPath != "" {
		return c.fullPath
	}
	return chi.RouteContext(c.Request.Context()).RoutePattern()
}

func (c *Context) JSON(code int, value interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Status(code)
	_ = json.NewEncoder(c.Writer).Encode(value)
}

func (c *Context) Data(code int, contentType string, data []byte) {
	c.Header("Content-Type", contentType)
	c.Status(code)
	_, _ = c.Writer.Write(data)
}

func (c *Context) String(code int, value string) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Status(code)
	_, _ = io.WriteString(c.Writer, value)
}

func (c *Context) Status(code int) { c.Writer.WriteHeader(code) }

func (c *Context) AbortWithStatusJSON(code int, value interface{}) {
	c.JSON(code, value)
	c.Abort()
}

func (c *Context) ShouldBindJSON(value interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(value)
}

func (c *Context) ShouldBindQuery(value interface{}) error {
	return bindQuery(c.Request.URL.Query(), value)
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, header, err := c.Request.FormFile(name)
	return header, err
}

type RouterGroup struct {
	router     chi.Router
	prefix     string
	middleware []HandlerFunc
}

type Engine struct{ *RouterGroup }

func New() *Engine {
	r := chi.NewRouter()
	g := &RouterGroup{router: r}
	return &Engine{RouterGroup: g}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) { e.router.ServeHTTP(w, r) }

// Chi exposes the underlying router for typed integrations such as Huma.
func (e *Engine) Chi() chi.Router { return e.router }

func (g *RouterGroup) Use(middleware ...HandlerFunc) {
	g.middleware = append(g.middleware, middleware...)
}

func (g *RouterGroup) Group(path string) *RouterGroup {
	return &RouterGroup{
		router:     g.router,
		prefix:     joinPath(g.prefix, path),
		middleware: append([]HandlerFunc{}, g.middleware...),
	}
}

func (g *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	g.handle(http.MethodGet, path, handlers...)
}
func (g *RouterGroup) POST(path string, handlers ...HandlerFunc) {
	g.handle(http.MethodPost, path, handlers...)
}
func (g *RouterGroup) PUT(path string, handlers ...HandlerFunc) {
	g.handle(http.MethodPut, path, handlers...)
}
func (g *RouterGroup) DELETE(path string, handlers ...HandlerFunc) {
	g.handle(http.MethodDelete, path, handlers...)
}
func (g *RouterGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	g.handle(http.MethodOptions, path, handlers...)
}

func (g *RouterGroup) handle(method, path string, handlers ...HandlerFunc) {
	all := append(append([]HandlerFunc{}, g.middleware...), handlers...)
	pattern := colonParams(joinPath(g.prefix, path))
	g.router.Method(method, pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Request:  r,
			Writer:   &ResponseWriter{ResponseWriter: w},
			index:    -1,
			handlers: all,
			fullPath: pattern,
		}
		ctx.Next()
	}))
}

func (g *RouterGroup) StaticFile(path, file string) {
	g.GET(path, func(c *Context) {
		http.ServeFile(c.Writer, c.Request, file)
	})
}

func WrapH(handler http.Handler) HandlerFunc {
	return func(c *Context) { handler.ServeHTTP(c.Writer, c.Request) }
}

func BasicAuth(accounts Accounts) HandlerFunc {
	return func(c *Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || accounts[user] != pass {
			c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
			c.JSON(http.StatusUnauthorized, H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CustomRecovery(handler func(*Context, interface{})) HandlerFunc {
	return func(c *Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				handler(c, recovered)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func CORS(origins, methods, headers []string) HandlerFunc {
	return func(c *Context) {
		origin := c.GetHeader("Origin")
		if contains(origins, "*") || contains(origins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			if contains(origins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			}
			c.Header("Access-Control-Allow-Methods", strings.Join(methods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(headers, ", "))
			c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")
		}
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}

func Gzip() HandlerFunc {
	return func(c *Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") || strings.Contains(c.GetHeader("Content-Type"), "image/") {
			c.Next()
			return
		}
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		writer := gzip.NewWriter(c.Writer.ResponseWriter)
		original := c.Writer.ResponseWriter
		c.Writer.ResponseWriter = writerResponseWriter{Writer: writer, original: original}
		c.Next()
		_ = writer.Close()
	}
}

type writerResponseWriter struct {
	*gzip.Writer
	original http.ResponseWriter
}

func (w writerResponseWriter) Header() http.Header { return w.original.Header() }

func (w writerResponseWriter) WriteHeader(code int) { w.original.WriteHeader(code) }

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func joinPath(prefix, path string) string {
	result := strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(path, "/")
	if result == "/" {
		return result
	}
	return strings.TrimRight(result, "/")
}

func colonParams(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

func bindQuery(values url.Values, target interface{}) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("query target must be a pointer to a struct")
	}
	rv = rv.Elem()
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if !field.CanSet() {
			continue
		}
		name := strings.Split(rt.Field(i).Tag.Get("form"), ",")[0]
		if name == "" {
			name = strings.ToLower(rt.Field(i).Name)
		}
		value := values.Get(name)
		if value == "" {
			continue
		}
		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(value, 10, field.Type().Bits())
			if err != nil {
				return err
			}
			field.SetInt(n)
		case reflect.Bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			field.SetBool(b)
		default:
			return fmt.Errorf("unsupported query field %s", rt.Field(i).Name)
		}
	}
	return nil
}
