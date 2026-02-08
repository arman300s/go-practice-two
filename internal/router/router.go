package router

import (
	"fmt"
	"net/http"
	"strings"
)

type Router struct {
	routes map[string]map[string]http.HandlerFunc // method -> path -> handler
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

func (r *Router) Handle(method, path string, handler http.HandlerFunc) {
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]http.HandlerFunc)
	}
	r.routes[method][path] = handler
}

func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodPost, path, handler)
}

func (r *Router) PATCH(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodPatch, path, handler)
}

func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodDelete, path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	if handlers, ok := r.routes[req.Method]; ok {
		if handler, ok := handlers[path]; ok {
			handler(w, req)
			return
		}
	}

	http.NotFound(w, req)
}

func (r *Router) PrintRoutes() {
	fmt.Println("Registered routes:")
	for method, paths := range r.routes {
		for path := range paths {
			fmt.Printf("  %s %s\n", method, path)
		}
	}
}
