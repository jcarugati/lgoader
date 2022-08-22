package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"
)

type ctrl struct {
	statusCode int
	response   interface{}
}

func (c *ctrl) mockHandler(w http.ResponseWriter, r *http.Request) {
	resp := []byte{}

	time.Sleep(500 * time.Millisecond)

	rt := reflect.TypeOf(c.response)
	if rt.Kind() == reflect.String {
		resp = []byte(c.response.(string))
	} else if rt.Kind() == reflect.Struct || rt.Kind() == reflect.Ptr {
		resp, _ = json.Marshal(c.response)
	}

	w.WriteHeader(c.statusCode)
	_, err := w.Write(resp)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func HttpMock(pattern string, statusCode int, response interface{}) *httptest.Server {
	c := &ctrl{statusCode, response}

	handler := http.NewServeMux()
	handler.HandleFunc(pattern, c.mockHandler)

	return httptest.NewServer(handler)
}
