package main

import (
	"net/http"
)

var (
	// HomeTemplate is the template for the home page
	HomeTemplate = Template{Files: []string{"main.html", "home.html"}}
	// SlackTemplate is the template for the slack page
	SlackTemplate = Template{Files: []string{"main.html", "slack.html"}}
)

// HomeHandler handles calls to the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	HomeTemplate.Execute(w, nil)
}

// SlackGetHandler handles GET calls to the slack page
func SlackGetHandler(w http.ResponseWriter, r *http.Request) {
	SlackTemplate.Execute(w, nil)
}

// SlackPostHandler processes the POSTed form and initiates an API call to Slack to have an
// invitation send to the provided email address.
func SlackPostHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form and get the value of the Email field
	r.ParseForm()
	email := r.Form.Get("Email")

	// Send the Slack invitation
	if err := SendSlackInvitation(email, token); err != nil {
		SlackTemplate.Execute(w, map[string]interface{}{"Alert": err.Error(), "Email": email})
		return
	}

	// All went well, execute the template with Success
	SlackTemplate.Execute(w, map[string]interface{}{"Success": true})
}
