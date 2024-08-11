package light_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	gohttp "github.com/hysem/go-http"
	"github.com/stretchr/testify/assert"
)

type ctxKey string

var (
	ctxKey1 = ctxKey("key1")
	ctxKey2 = ctxKey("key2")
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Fprintf(w,
		"[%s] %s => ctx_key1: %v; ctx_key2: %v",
		r.Method,
		r.URL.String(),
		ctx.Value(ctxKey1),
		ctx.Value(ctxKey2),
	)
}

func middlewareWithKey(key ctxKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			v, _ := ctx.Value(key).(int)
			v++
			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, key, v)))
		}
		return http.HandlerFunc(fn)
	}
}
func TestRouter_HttpMethods(t *testing.T) {
	r := gohttp.NewRouter()
	r.Connect("/user", handler)
	r.Delete("/user", handler)
	r.Get("/user", handler)
	r.Head("/user", handler)
	r.Options("/user", handler)
	r.Patch("/user", handler)
	r.Post("/user", handler)
	r.Put("/user", handler)
	r.Trace("/user", handler)

	testCases := []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}

	for _, tc := range testCases {
		method := tc
		t.Run(fmt.Sprintf("method: %s", method), func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(method, "/user", nil)

			r.ServeHTTP(rec, req)

			expectedBody := fmt.Sprintf("[%s] /user => ctx_key1: <nil>; ctx_key2: <nil>", method)

			assert.Equal(t, expectedBody, rec.Body.String(), "response body mismatch")
		})
	}
}

func TestRouter_With(t *testing.T) {
	mw1 := middlewareWithKey(ctxKey1)
	mw2 := middlewareWithKey(ctxKey2)

	r := gohttp.NewRouter()
	r.With(mw1, mw2).Get("/v1/health", handler)
	r.With(mw2).Get("/v2/health", handler)

	testCases := map[string]struct {
		endpoint     string
		expectedBody string
	}{
		`with two middlewares`: {
			endpoint:     "/v1/health",
			expectedBody: "[GET] /v1/health => ctx_key1: 1; ctx_key2: 1",
		},
		`with one middleware`: {
			endpoint:     "/v2/health",
			expectedBody: "[GET] /v2/health => ctx_key1: <nil>; ctx_key2: 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedBody, rec.Body.String(), "response body mismatch")
		})
	}
}

func TestRouter_Use(t *testing.T) {
	mw1 := middlewareWithKey(ctxKey1)
	mw2 := middlewareWithKey(ctxKey2)

	r := gohttp.NewRouter()
	r.Get("/v1/health", handler)
	r.Use(mw1, mw2)
	r.Get("/v2/health", handler)

	testCases := map[string]struct {
		endpoint     string
		expectedBody string
	}{
		`with no middlewares`: {
			endpoint:     "/v1/health",
			expectedBody: "[GET] /v1/health => ctx_key1: <nil>; ctx_key2: <nil>",
		},
		`with middlewares`: {
			endpoint:     "/v2/health",
			expectedBody: "[GET] /v2/health => ctx_key1: 1; ctx_key2: 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedBody, rec.Body.String(), "response body mismatch")
		})
	}
}

func TestRouter_Group(t *testing.T) {
	mw1 := middlewareWithKey(ctxKey1)
	mw2 := middlewareWithKey(ctxKey2)

	r := gohttp.NewRouter()

	r.Use(mw1, mw2)

	group1 := r.Group(func(r gohttp.Router) {
		r.Use(mw1)
		r.Get("/v1/health", handler)
	})

	group1.Get("/v2/health", handler)

	r.Get("/v3/health", handler)

	testCases := map[string]struct {
		endpoint     string
		expectedBody string
	}{
		`within group1`: {
			endpoint:     "/v1/health",
			expectedBody: "[GET] /v1/health => ctx_key1: 2; ctx_key2: 1",
		},
		`on group1`: {
			endpoint:     "/v2/health",
			expectedBody: "[GET] /v2/health => ctx_key1: 2; ctx_key2: 1",
		},
		`outside group1`: {
			endpoint:     "/v3/health",
			expectedBody: "[GET] /v3/health => ctx_key1: 1; ctx_key2: 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedBody, rec.Body.String(), "response body mismatch")
		})
	}
}

func TestRouter_Route(t *testing.T) {
	mw1 := middlewareWithKey(ctxKey1)
	mw2 := middlewareWithKey(ctxKey2)

	r := gohttp.NewRouter()

	r.Use(mw1, mw2)

	v1Route := r.Route("/v1", func(r gohttp.Router) {
		r.Use(mw1)
		r.Get("/health", handler)
	})

	v1Route.Get("/ping", handler)

	testCases := map[string]struct {
		endpoint     string
		expectedBody string
	}{
		`within group1`: {
			endpoint:     "/v1/health",
			expectedBody: "[GET] /v1/health => ctx_key1: 2; ctx_key2: 1",
		},
		`on group1`: {
			endpoint:     "/v1/ping",
			expectedBody: "[GET] /v1/ping => ctx_key1: 2; ctx_key2: 1",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedBody, rec.Body.String(), "response body mismatch")
		})
	}
}
