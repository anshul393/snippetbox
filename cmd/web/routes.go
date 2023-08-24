package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Create a handler function which wraps our notFound() helper, and then
	// assign it as the custom handler for 404 Not Found responses. You can also
	// set a custom handler for 405 Method Not Allowed responses by setting
	// router.MethodNotAllowed in the same way too.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf)
	protected := dynamic.Append(app.requireAuthentication)
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))

	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
	router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))
	router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.accountView))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}

//
