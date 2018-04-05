package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
)

const (
	URL = "https://golangbg.slack.com/api/users.admin.invite"
)

var token string

type Template struct {
	Files []string
	tpl   *template.Template
	once  sync.Once
}

func (t *Template) Execute(w http.ResponseWriter, data map[string]interface{}) {
	t.once.Do(func() {
		tpl, err := template.ParseFiles(t.Files...)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		t.tpl = tpl
	})
	if t.tpl != nil {
		t.tpl.ExecuteTemplate(w, "main", data)
	}
}

var (
	HomeTemplate  = Template{Files: []string{"templates/main.html", "templates/home.html"}}
	SlackTemplate = Template{Files: []string{"templates/main.html", "templates/slack.html"}}
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	HomeTemplate.Execute(w, nil)
}

func SlackGetHandler(w http.ResponseWriter, r *http.Request) {
	SlackTemplate.Execute(w, nil)
}

type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func SlackPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("Email")
	client := http.Client{}

	req, _ := http.NewRequest("POST", URL, nil)
	q := req.URL.Query()
	q.Set("email", email)
	q.Set("token", token)
	q.Set("set_active", "true")
	req.URL.RawQuery = q.Encode()

	a, err := client.Do(req)
	if err != nil {
		SlackTemplate.Execute(w, map[string]interface{}{"Alert": err.Error(), "Email": email})
		return
	}
	defer a.Body.Close()
	if a.StatusCode != http.StatusOK {
		SlackTemplate.Execute(w, map[string]interface{}{"Alert": "Something went wrong: " + a.Status, "Email": email})
		return
	}

	var resp Response
	if err := json.NewDecoder(a.Body).Decode(&resp); err != nil {
		SlackTemplate.Execute(w, map[string]interface{}{"Alert": err.Error(), "Email": email})
		return
	}

	if !resp.OK {
		var err string
		switch resp.Error {
		case "already_invited":
			err = "User has already received an email invitation"
		case "already_in_team":
			err = "User is already part of the team"
		case "channel_not_found":
			err = "Provided channel ID does not match a real channel"
		case "sent_recently":
			err = "Email has been sent recently already"
		case "user_disabled":
			err = "User account has been deactivated"
		case "missing_scope":
			err = "Not authorized for 'client' scope"
		case "invalid_email":
			err = "Invalid email address"
		case "not_allowed":
			err = "Not allowed, SSO is enabeld"
		case "not_allowed_token_type":
			err = "Token type is invalid"
		default:
			err = "Unknown error: " + resp.Error
		}
		SlackTemplate.Execute(w, map[string]interface{}{"Alert": err, "Email": email})
		return
	}

	SlackTemplate.Execute(w, map[string]interface{}{"Success": true})
}

func main() {
	token = os.Getenv("SI_TOKEN")
	if token == "" {
		log.Fatal("SI_TOKEN is empty")
	}

	addr := os.Getenv("SI_PORT")
	if addr == "" {
		addr = ":80"
	}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.HandleFunc("/slack", SlackGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/slack", SlackPostHandler).Methods(http.MethodPost)
	r.HandleFunc("/", HomeHandler)

	log.Fatal(http.ListenAndServe(addr, r))
}
