// Writing a basic HTTP server is easy using the
// `net/http` package.
package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	key   = []byte("d78adfdf210f3bfed2647b2529864266b53d4a7f5543f9712d44d4e3aa72fa9c")
	store = sessions.NewCookieStore(key)
	tmpl  = template.Must(template.ParseGlob("templates/*.html"))
	users = make(map[string]User)
)

type User struct {
	Lastname    string
	Firstname   string
	Email       string
	Phone       string
	Picture     string
	Password    string
	HomeAddress string
	WorkAddress string
}

func renderHome(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "session")
	auth, exists := session.Values["user"].(string)
	if !exists {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	user, exists := users[auth]
	if !exists {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	if user.HomeAddress == "" || user.WorkAddress == "" {
		tmpl.ExecuteTemplate(w, "initProfile.html", struct {
			User User
		}{
			User: user,
		})
		return
	}
	tmpl.ExecuteTemplate(w, "index.html", struct {
		User User
	}{
		User: user,
	})
}

func renderLogin(w http.ResponseWriter, req *http.Request) {
	tmpl.ExecuteTemplate(w, "login.html", nil)
}
func renderRegister(w http.ResponseWriter, req *http.Request) {
	tmpl.ExecuteTemplate(w, "register.html", nil)
}

func login(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "session")

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := req.FormValue("email")
	password := req.FormValue("password")
	user := users[email]

	if user.Password != password {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	session.Values["user"] = user.Email
	session.Save(req, w)

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
func logout(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "session")
	session.Values["user"] = nil
	session.Options.MaxAge = -1
	session.Save(req, w)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
func register(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	user := User{
		Lastname:  req.FormValue("nom"),
		Firstname: req.FormValue("prenom"),
		Email:     req.FormValue("email"),
		Phone:     req.FormValue("tel"),
		Picture:   req.FormValue("photo"),
		Password:  req.FormValue("password"),
	}

	users[user.Email] = user

	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

func setAdresses(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "session")
	auth, exists := session.Values["user"].(string)
	if !exists {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	user, exists := users[auth]
	if !exists {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	user.HomeAddress = req.FormValue("home")
	user.WorkAddress = req.FormValue("work")
	users[auth] = user
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", renderHome).Methods("GET")
	r.HandleFunc("/register", renderRegister).Methods("GET")
	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/login", renderLogin).Methods("GET")
	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("GET")
	r.HandleFunc("/address", setAdresses).Methods("POST")

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))
	http.ListenAndServe(":8080", r)
}
