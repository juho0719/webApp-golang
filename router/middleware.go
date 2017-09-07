package main

import (
	"time"
	"log"
	"net/http"
	"fmt"
	"encoding/json"
	"strings"
	"path"
)

type Middleware func(next HandlerFunc) HandlerFunc

func logHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		t := time.Now()

		next(c)

		log.Printf("[%s] %q %v\n",
			c.Request.Method,
			c.Request.URL.String(),
			time.Now().Sub(t))
	}
}

func recoverHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(c.ResponseWriter,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
			}
		}()
		next(c)
	}
}

func parseFormHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		c.Request.ParseForm()
		fmt.Println(c.Request.PostForm)
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				c.Params[k] = v[0]
			}
		}
		next(c)
	}
}

func parseJsonBodyHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		var m map[string]interface{}
		if json.NewDecoder(c.Request.Body).Decode(&m); len(m) > 0 {
			for k, v := range m {
				c.Params[k] = v
			}
		}
		next(c)
	}
}

func staticHandler(next HandlerFunc) HandlerFunc {
	var (
		dir = http.Dir(".")
		indexFile = "index.html"
	)

	return func(c *Context) {
		//http메소드가 GET이나 HEAD가 아니면 다음 핸들러 수행
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			next(c)
			return
		}

		file := c.Request.URL.Path
		//URL경로에 해당하는 파일 열기
		f, err := dir.Open(file)
		if err != nil {
			//파일 열기 실패 후 다음 핸들러 수행
			next(c)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			//파일 상태가 정상이 아니면 다음 핸들러 수행
			next(c)
			return
		}

		if fi.IsDir() {
			//디렉토리 경로를 URL로 사용하면 끝에 "/"붙임
			if !strings.HasSuffix(c.Request.URL.Path, "/") {
				http.Redirect(c.ResponseWriter, c.Request, c.Request.URL.Path+"/", http.StatusFound)
				return
			}

			//디렉토리를 가리키는 URL경로에 indexFile이름을 붙여 전체 파일 경로 생성
			file = path.Join(file, indexFile)

			//indexFile 열기
			f, err := dir.Open(file)
			if err != nil {
				next(c)
				return
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil || fi.IsDir(){
				//indexFile 상태가 정상이 아니면 다음 핸들러 수행
				next(c)
				return
			}
		}

		//file의 내용전달
		http.ServeContent(c.ResponseWriter, c.Request, file, fi.ModTime(), f)
	}
}