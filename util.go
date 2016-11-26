package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// serveTemplate parses template files and serves resulting html.
func serveTemplate(w http.ResponseWriter, data Data, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}

	t := template.Must(template.ParseFiles(files...))
	err := t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

// getSession returns the current session if it exists, otherwise an empty
// session is returned.
func getSession(r *http.Request) Session {
	cUser, err := r.Cookie("session-id")
	var session Session
	if err == nil {
		session.Id = cUser.Value
		session.LoggedIn = true
	}
	cQuestion, err := r.Cookie("current-question")
	if err == nil {
		curQuestion, _ := strconv.Atoi(cQuestion.Value)
		session.CurQuestion = curQuestion
	}
	qIndex, err := r.Cookie("current-qindex")
	if err == nil {
		qIndex, _ := strconv.Atoi(qIndex.Value)
		session.QuestionIndex = qIndex
	}
	return session
}

// updateUserResponses fetches the user with id username and adds response to
// the user's list of responses.
// It returns true if update was successful, false otherwise.
func updateUserResponses(w http.ResponseWriter, r *http.Request, username string, response string) bool {
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", username, 0, nil)
	var user User
	err := datastore.Get(ctx, key, &user)
	if err != nil {
		return false
	}
	qRes, _ := strconv.Atoi(response)
	user.Responses = append(user.Responses, qRes)
	if len(user.Responses) == MAX_QUESTIONS {
		user.SurveyComplete = true
	}
	_, err = datastore.Put(ctx, key, &user)
	if err != nil {
		return false
	}
	return true
}
