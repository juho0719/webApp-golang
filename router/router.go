package router

import (
	"net/http"
	"strings"
)

type router struct {
	handlers map[string]map[string]http.HandlerFunc
}

func (r *router) HandleFunc(method, pattern string, h http.HandlerFunc) {
	m, ok := r.handlers[method]
	if !ok {
		m = make(map[string]http.HandlerFunc)
		r.handlers[method] = m
	}

	m[pattern] = h
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if m, ok := r.handlers[req.Method]; ok {
		if h, ok := m[req.URL.Path]; ok {
			h(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

func match(pattern, path string) (bool, map[string]string) {
	if pattern == path {
		return true, nil
	}

	patterns := strings.Split(pattern, "/")
	paths := strings.Split(path, "/")

	if len(len(patterns) != len(paths)) {
		return false, nil
	}

	params := make(map[string]string)

	for i:=0; i<len(patterns); i++ {
		switch {
		case patterns[i] == paths[i]:
		case len(patterns) == paths[i]:
		case len(patterns[i])>0 && patterns[i][0] == ':':
			params[patterns[i][1:]] = paths[i]
		default:
			return false, nil
		}
	}

	return true, params
}