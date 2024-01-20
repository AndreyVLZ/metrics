package handlers

import "net/http"

type Handlers interface {
	// [ / ]
	ListHandler(http.ResponseWriter, *http.Request)
	// [ /ping ]
	PingHandler(http.ResponseWriter, *http.Request)
	// [ /value/ ]
	ValueHandler
	// [ /update/ ]
	UpdateHandler
}

type ValueHandler interface {
	// [ GET ]
	GetValueHandler() http.Handler
	// [ POST ]
	PostValueHandler() http.Handler
}

type UpdateHandler interface {
	// [ POST ]
	PostUpdateHandler() http.Handler
}
