package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "github.com/anshul393/snippetbox/db/sqlc"
	"github.com/anshul393/snippetbox/internal/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type userSignUpForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	app.infoLog.Print(r.Context())

	snippets, err := app.dtb.GetLatestSnippets(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)

	data.Snippets = snippets

	app.render(w, http.StatusOK, "home.tmpl.html", data)

}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))

	// id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		app.notFound(w)
		return
	}

	snippet, err := app.dtb.GetSnippet(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.infoLog.Print(r.Context())

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, http.StatusOK, "view.tmpl.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusOK, "create.tmpl.html", data)
		return
	}

	data := db.InsertSnippetParams{
		Title:   form.Title,
		Content: form.Content,
		Column3: form.Expires,
	}

	id, err := app.dtb.InsertSnippet(r.Context(), data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Form will be processed. Please keep patience"))
	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{Expires: 365}
	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignUpForm{}
	app.render(w, http.StatusOK, "signup.tmpl.html", data)

	fmt.Fprintln(w, "Display a HTML form for signing up a new user...")
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := userSignUpForm{}
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using our helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(form.Password), 12)
	if err != nil {
		app.serverError(w, err)
		return
	}

	UserInsertParam := db.InsertUserParams{
		Name:           form.Name,
		Email:          form.Email,
		HashedPassword: string(hash),
		Created:        time.Now(),
	}

	err = app.dtb.InsertUser(r.Context(), UserInsertParam)
	if err != nil {
		var PSQLError *pq.Error
		if errors.As(err, &PSQLError) {
			if PSQLError.Code == "23505" && strings.Contains(PSQLError.Message, "users_uc_email") {
				form.AddFieldError("email", "Email address is already in use")
				data := app.newTemplateData(r)
				data.Form = form
				app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
			} else {
				app.serverError(w, err)
			}
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl.html", data)
}
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := userLoginForm{}
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	// Authenticate the user on the basis of its email and password

	id, err := app.Authenticate(r, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl.html", data)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	path := app.sessionManager.PopString(r.Context(), "path")

	app.sessionManager.Put(r.Context(), "email", form.Email)
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	app.sessionManager.Put(r.Context(), "flash", "You have successfully logged in...")
	http.Redirect(w, r, path, http.StatusSeeOther)

}
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.tmpl.html", data)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	id := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
	user, _ := app.dtb.GetUserByID(r.Context(), int32(id))
	u := User{Name: user.Name, Email: user.Email, Joined: user.Created}
	data := app.newTemplateData(r)

	data.User = &u

	app.render(w, http.StatusOK, "account.tmpl.html", data)

}
