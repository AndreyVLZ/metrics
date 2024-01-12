package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

var ErrOnlyPostRequest error = errors.New("only Post")

type mwMethod struct {
	http.Handler
	methodStr string
}

func (m *mwMethod) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if m.methodStr != req.Method {
		http.Error(rw, fmt.Sprintf("method is not %v\n", m.methodStr), http.StatusBadRequest)
		return
	}

	m.Handler.ServeHTTP(rw, req)
}

func (m *mwMethod) method() string { return m.methodStr }

func Get(next http.Handler) *mwMethod {
	return method(http.MethodGet, next)
}

func Post(next http.Handler) *mwMethod {
	return method(http.MethodPost, next)
}

func method(method string, next http.Handler) *mwMethod {
	return &mwMethod{Handler: next, methodStr: method}
}

type mwMethods struct {
	methods map[string]http.Handler
}

func Methods(arr ...*mwMethod) *mwMethods {
	mwMethods := newMethods()
	for i := range arr {
		mwMethods.addMethod(arr[i])
	}

	return mwMethods
}

func newMethods() *mwMethods {
	return &mwMethods{
		methods: make(map[string]http.Handler, 9),
	}
}

func (ms *mwMethods) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h, ok := ms.methods[req.Method]
	if !ok {
		http.Error(rw, "method not support", http.StatusBadRequest)
		return
	}

	h.ServeHTTP(rw, req)
}

func (ms *mwMethods) addMethod(mwMethod *mwMethod) {
	methodStr := mwMethod.method()
	_, ok := ms.methods[methodStr]
	if ok {
		log.Printf("method %v is exists", methodStr)
		return
	}

	ms.methods[methodStr] = mwMethod
}

func Method(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, ErrOnlyPostRequest.Error(), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(rw, req)
	}
}
