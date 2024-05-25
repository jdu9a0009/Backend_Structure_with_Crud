package web

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"syscall"
	"time"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values are stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// A Handler is a type that handles a http request within our own little mini framework.
type Handler func(c *Context) error

// registered keeps track of handlers registered to the http default server
// mux. This is a singleton and used by the standard library for metrics
// and profiling. The application may want to add other handlers like
// readiness and liveness to that mux. If this is not tracked, the routes
// could try to be registered more than once, causing panic.
var registered = make(map[string]bool)

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct
type App struct {
	*gin.Engine
	shutdown    chan os.Signal
	mw          []Middleware
	DefaultLang string
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, defaultLang string, mw ...Middleware) *App {
	engine := gin.Default()

	// cors
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	//engine.Static("/media", "./media")

	return &App{
		Engine:      engine,
		shutdown:    shutdown,
		mw:          mw,
		DefaultLang: defaultLang,
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// handle performs the real work of applying boilerplate and framework code
// for handler.
func (a *App) handle(debug bool, method string, path string, handler Handler, mw ...Middleware) {
	if debug {
		// Track all the handlers that are being registered so we don't have
		// the same handlers registered twice to this singleton.
		if exists := registered[method+path]; exists {
			return
		}
		registered[method+path] = true
	}

	// First wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to handler chain.
	handler = wrapMiddleware(a.mw, handler)

	// The function execute for each request.
	h := func(c *gin.Context) {

		//defer func() {
		//	if r := recover(); r != nil {
		//		c.JSON(http.StatusInternalServerError, map[string]interface{}{
		//			"error":  fmt.Sprintf("Recovered error: %v", r),
		//			"status": false,
		//		})
		//	}
		//}()

		// ###########################################
		//// Start or expand a distributed trace.
		ctx := c.Request.Context()
		//ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, r.URL.Path)
		//defer span.End()
		// ###########################################

		// Set the context with required values to
		// process the request.
		v := Values{
			//TraceID: span.SpanContext().TraceID.String(),
			Now: time.Now(),
		}

		lang := a.DefaultLang

		if len(c.Request.Header["Accept-Language"]) > 0 {
			lang = c.Request.Header["Accept-Language"][0]
		}

		ctx = context.WithValue(ctx, KeyValues, &v)
		ctx = context.WithValue(ctx, "lang", lang)
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		webContext := NewContext(c, ctx)
		if err := handler(webContext); err != nil {
			a.SignalShutdown()
			return
		}
	}

	a.Handle(method, path, h)

}

// HandleFunc sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) HandleFunc(method string, path string, handler Handler, mw ...Middleware) {
	a.handle(false, method, path, handler, mw...)
}

func (a *App) Get(path string, handler Handler, mw ...Middleware) {
	a.HandleFunc(http.MethodGet, path, handler, mw...)
}

func (a *App) Post(path string, handler Handler, mw ...Middleware) {
	a.HandleFunc(http.MethodPost, path, handler, mw...)
}

func (a *App) Put(path string, handler Handler, mw ...Middleware) {
	a.HandleFunc(http.MethodPut, path, handler, mw...)
}

func (a *App) Patch(path string, handler Handler, mw ...Middleware) {
	a.HandleFunc(http.MethodPatch, path, handler, mw...)
}

func (a *App) Delete(path string, handler Handler, mw ...Middleware) {
	a.HandleFunc(http.MethodDelete, path, handler, mw...)
}

func (a *App) GroupFunc(path string) *App {
	return a
}
