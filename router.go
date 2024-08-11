package light

import (
	"fmt"
	"net/http"
)

// Router consisting of the core routing methods used by chi's Mux,
// using only the standard net/http.
type Router interface {
	http.Handler

	// Use appends one or more middlewares onto the Router stack.
	Use(middlewares ...func(http.Handler) http.Handler)

	// With adds inline middlewares for an endpoint handler.
	With(middlewares ...func(http.Handler) http.Handler) Router

	// Group adds a new inline-Router along the current routing
	// path, with a fresh middleware stack for the inline-Router.
	Group(fn func(r Router)) Router

	// Route mounts a sub-Router along a `pattern`` string.
	Route(pattern string, fn func(r Router)) Router

	// Method and MethodFunc adds routes for `pattern` that matches
	// the `method` HTTP method.
	Method(method, pattern string, h http.Handler)
	MethodFunc(method, pattern string, h http.HandlerFunc)

	// // HTTP-method routing along `pattern`
	Connect(pattern string, h http.HandlerFunc)
	Delete(pattern string, h http.HandlerFunc)
	Get(pattern string, h http.HandlerFunc)
	Head(pattern string, h http.HandlerFunc)
	Options(pattern string, h http.HandlerFunc)
	Patch(pattern string, h http.HandlerFunc)
	Post(pattern string, h http.HandlerFunc)
	Put(pattern string, h http.HandlerFunc)
	Trace(pattern string, h http.HandlerFunc)
}

type routerContext struct {
	pattern string
	*http.ServeMux
	middlewares []func(next http.Handler) http.Handler
}

func NewRouter() Router {
	return &routerContext{
		ServeMux: http.NewServeMux(),
	}
}

func (r *routerContext) Use(middlewares ...func(next http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *routerContext) With(middlewares ...func(next http.Handler) http.Handler) Router {
	nr := &routerContext{
		ServeMux: r.ServeMux,
		pattern:  r.pattern,
	}
	nr.middlewares = append(nr.middlewares, r.middlewares...)
	nr.middlewares = append(nr.middlewares, middlewares...)
	return nr
}

func (r *routerContext) Group(fn func(r Router)) Router {
	nr := &routerContext{
		ServeMux: r.ServeMux,
		pattern:  r.pattern,
	}
	nr.middlewares = append(nr.middlewares, r.middlewares...)
	fn(nr)
	return nr
}

func (r *routerContext) Route(pattern string, fn func(r Router)) Router {
	nr := &routerContext{
		ServeMux: r.ServeMux,
		pattern:  fmt.Sprintf("%s%s", r.pattern, pattern),
	}
	nr.middlewares = append(nr.middlewares, r.middlewares...)
	fn(nr)
	return nr
}

func (r *routerContext) Method(method, pattern string, h http.Handler) {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}

	if method != "" {
		method = fmt.Sprintf("%s ", method)
	}

	r.Handle(fmt.Sprintf("%s%s%s", method, r.pattern, pattern), h)
}

func (r *routerContext) MethodFunc(method, pattern string, h http.HandlerFunc) {
	r.Method(method, pattern, h)
}

func (r *routerContext) Connect(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodConnect, pattern, h)
}

func (r *routerContext) Delete(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodDelete, pattern, h)
}

func (r *routerContext) Get(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodGet, pattern, h)
}

func (r *routerContext) Head(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodHead, pattern, h)
}

func (r *routerContext) Options(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodOptions, pattern, h)
}

func (r *routerContext) Patch(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodPatch, pattern, h)
}

func (r *routerContext) Post(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodPost, pattern, h)
}

func (r *routerContext) Put(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodPut, pattern, h)
}

func (r *routerContext) Trace(pattern string, h http.HandlerFunc) {
	r.MethodFunc(http.MethodTrace, pattern, h)
}
