package handlers

import "net/http"

type Handlers interface {
	// [ / ]
	ListHandler(http.ResponseWriter, *http.Request)
	// [ /ping ]
	PingHandler() http.Handler
	// [ /value/ ]
	ValueHandler
	// [ /update/ ]
	UpdateHandler
	// [ /updates/ ]
	PostUpdatesHandler() http.Handler
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
