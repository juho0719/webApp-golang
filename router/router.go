package main

import (
	"net/http"
	"strings"
)

type router struct {
	handlers map[string]map[string]HandlerFunc
}

func (r *router) HandleFunc(method, pattern string, h HandlerFunc) {
	m, ok := r.handlers[method]
	if !ok {
		m = make(map[string]HandlerFunc)
		r.handlers[method] = m
	}

	m[pattern] = h
}

func match(pattern, path string) (bool, map[string]string) {
	if pattern == path {
		return true, nil
	}

	patterns := strings.Split(pattern, "/")
	paths := strings.Split(path, "/")

	if len(patterns) != len(paths) {
		return false, nil
	}

	params := make(map[string]string)

	for i:=0; i<len(patterns); i++ {
		switch {
		case patterns[i] == paths[i]:
		case len(patterns[i])>0 && patterns[i][0] == ':':
			params[patterns[i][1:]] = paths[i]
		default:
			return false, nil
		}
	}

	return true, params
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for pattern, handler := range r.handlers[req.Method] {
		if ok, params := match(pattern, req.URL.Path); ok {
			//Context
			c := Context{
				Params: make(map[string]interface{}),
				ResponseWriter: w,
				Request: req,
			}
			for k, v := range params {
				c.Params[k] = v
			}
			// 요청 url에 해당하는 handler 수행
			handler(&c)
			return
		}
	}

	http.NotFound(w, req)
	return
}

func (r *router) handler() HandlerFunc {
	return func(c *Context) {
		//http메소드에 맞는 모든 handlers를 반복하며 요청 URL에 해당하는 핸들러를 찾음
		for pattern, handler := range r.handlers[c.Request.Method] {
			if ok, params := match(pattern, c.Request.URL.Path); ok {
				for k, v := range params {
					c.Params[k]= v
				}
				//요청 URL에 해당하는 핸들러 수행
				handler(c)
				return
			}
		}

		//요청 URL에 해당하는 핸들러를 찾지 못하면 NotFound
		http.NotFound(c.ResponseWriter, c.Request)
		return
	}
}