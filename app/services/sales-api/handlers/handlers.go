// Package handlers manages the different versions of the API.
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/debug/checkgrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/testgrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/usergrp"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/web/mid"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(shutdown chan os.Signal, log *zap.SugaredLogger, a *auth.Auth, db *sqlx.DB) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Error(log), mid.Metrics(), mid.Panics())

	app.Handle(http.MethodGet, "/test", testgrp.Handler)
	app.Handle(http.MethodGet, "/testauth", testgrp.Handler, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))

	authen := mid.Authenticate(a)
	admin := mid.Authorize(auth.RoleAdmin)

	// Register user management and authentication endpoints.
	ugh := usergrp.Handlers{
		User: user.NewCore(log, db),
		Auth: a,
	}
	app.Handle(http.MethodGet, "/users/token", ugh.Token)
	app.Handle(http.MethodGet, "/users/:page/:rows", ugh.Query, authen, admin)
	app.Handle(http.MethodGet, "/users/:id", ugh.QueryByID, authen)
	app.Handle(http.MethodPost, "/users", ugh.Create, authen, admin)
	app.Handle(http.MethodPut, "/users/:id", ugh.Update, authen, admin)
	app.Handle(http.MethodDelete, "/users/:id", ugh.Delete, authen, admin)

	return app
}

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the
// DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// DebugMux registers all the debug standard library routes and then custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk since
// a dependency could inject a handler into our service without us knowing it.
func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoints.
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}
	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}
