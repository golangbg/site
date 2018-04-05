package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response is the response returned by Slack after an API call to send an invitation
type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// SendSlackInvitation calls the Slack API to have an invitation sent to an email address
// Arguments are the email address for the delivery of the invitation and the slack API token
func SendSlackInvitation(email, token string) error {
	// Create a http client, request and set the required parameters for the API call to slack
	client := http.Client{}
	req, _ := http.NewRequest("POST", URL, nil)
	q := req.URL.Query()
	q.Set("email", email)
	q.Set("token", token)
	q.Set("set_active", "true")
	req.URL.RawQuery = q.Encode()

	// Make the API call
	a, err := client.Do(req)
	if err != nil {
		return err
	}
	// Ensure the Body will be closed
	defer a.Body.Close()

	// If we got another StatusCode than 200, show the error on the ResponseWriter
	if a.StatusCode != http.StatusOK {
		return fmt.Errorf(a.Status)
	}

	// Decode the response from JSON to Go
	var resp Response
	if err := json.NewDecoder(a.Body).Decode(&resp); err != nil {
		// Encoding went wrong, show the error on the ResponseWriter
		// Normally this shouldn't happen, but just in case
		return err
	}

	// If resp.OK isn't true we got an API error, determine which error it was and show it to the user
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
		return fmt.Errorf(err)
	}
	return nil
}
