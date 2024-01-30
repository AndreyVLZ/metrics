package handlers

import "net/http"

type Handlers interface {
	// [ / ]
	ListHandler() http.Handler
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
