package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const MAX_QUESTIONS = 4
const MAX_ANSWERS = 4

var (
	questions = []string{
		"How do you feel today?",
		"Given your mood today, which of the following places would you prefer to be right now?",
		"Which of the following makes you anxious?",
		"Which of the following do you need most in your life right now?",
	}
	answers = [][]string{
		{"Bored", "Excited", "Happy", "Sad"},
		{"Abandoned Farm", "Woods", "Busy city", "Sea"},
		{"Friends", "Family", "Strangers", "Authorities"},
		{"Friends", "Money", "Career Advancement", "Vacation"},
	}
)

// main the server main function.
func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/about", about)
	http.HandleFunc("/createuser", createUser)
	http.HandleFunc("/dashboard", dashboard)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/survey", handleSurvey)
	http.HandleFunc("/api/recordUserResponse", recordUserResponse)
	http.HandleFunc("/api/aggregateResponses", aggregateResponses)
	appengine.Main()
}

// GET /
// home serves the home page.
func home(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	data := Data{
		Session: session,
	}
	ctx := appengine.NewContext(r)
	var user User
	if session.LoggedIn {
		key := datastore.NewKey(ctx, "User", session.Id, 0, nil)
		err := datastore.Get(ctx, key, &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if user.SurveyComplete {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	serveTemplate(w, data, "layout", "navbar", "login", "register", "landing", "footer")
}

// GET /about
// about serves the info page.
func about(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	data := Data{
		Session: session,
	}
	serveTemplate(w, data, "layout", "navbar", "login", "register", "about", "footer")
}

// GET /dashboard
// dashboard serves dashboard page containing user's survey responses and
// charts showing the distribution of responses to each survey question.
func dashboard(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	data := Data{
		Session:   session,
		Questions: questions,
	}
	if !session.LoggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ctx := appengine.NewContext(r)
	var user User
	if session.LoggedIn {
		key := datastore.NewKey(ctx, "User", session.Id, 0, nil)
		err := datastore.Get(ctx, key, &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if !user.SurveyComplete {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	responses := []string{}
	for i := 0; i < len(user.Responses); i++ {
		responses = append(responses, answers[i][user.Responses[i]])
	}
	data.Responses = responses
	serveTemplate(w, data, "layout", "navbar", "login", "register", "dashboard", "footer")
}

// POST /createuser
// createUser adds user to database.
func createUser(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	username := r.FormValue("username")
	password := r.FormValue("password")
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := User{
		Id:             username,
		Password:       string(hash[:]),
		Responses:      []int{},
		SurveyComplete: false,
	}
	key := datastore.NewKey(ctx, "User", user.Id, 0, nil)
	err := datastore.Get(ctx, key, &user)
	if err == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	_, err = datastore.Put(ctx, key, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cUser := &http.Cookie{
		Name:  "session-id",
		Value: user.Id,
	}
	cQuestion := &http.Cookie{
		Name:  "current-question",
		Value: strconv.Itoa(1),
	}
	qIndex := &http.Cookie{
		Name:  "current-qindex",
		Value: strconv.Itoa(0),
	}
	http.SetCookie(w, cUser)
	http.SetCookie(w, cQuestion)
	http.SetCookie(w, qIndex)
	http.Redirect(w, r, "/", http.StatusFound)
}

// POST /login
// login authenticates user.
func login(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	username := r.FormValue("username")
	key := datastore.NewKey(ctx, "User", username, 0, nil)
	var user User
	err := datastore.Get(ctx, key, &user)
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	password := r.FormValue("password")
	hash := []byte(user.Password)
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	cUser := &http.Cookie{
		Name:  "session-id",
		Value: user.Id,
	}
	questionIndex := 0
	curQuestion := 1
	if user.SurveyComplete {
		curQuestion = -1
		questionIndex = -1
	} else {
		user.Responses = []int{}
		_, err = datastore.Put(ctx, key, &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	cQuestion := &http.Cookie{
		Name:  "current-question",
		Value: strconv.Itoa(curQuestion),
	}
	qIndex := &http.Cookie{
		Name:  "current-qindex",
		Value: strconv.Itoa(questionIndex),
	}
	http.SetCookie(w, cUser)
	http.SetCookie(w, cQuestion)
	http.SetCookie(w, qIndex)
	w.Write([]byte("true"))
}

// POST /logout
// logout deletes the session data and redirects user back to home.
func logout(w http.ResponseWriter, r *http.Request) {
	cUser, _ := r.Cookie("session-id")
	cUser = &http.Cookie{
		Name:   "session-id",
		Value:  "",
		MaxAge: -1,
	}
	cQuestion, _ := r.Cookie("current-question")
	cQuestion = &http.Cookie{
		Name:   "current-question",
		Value:  "",
		MaxAge: -1,
	}
	qIndex, _ := r.Cookie("current-qindex")
	qIndex = &http.Cookie{
		Name:   "current-qindex",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cUser)
	http.SetCookie(w, cQuestion)
	http.SetCookie(w, qIndex)
	http.Redirect(w, r, "/", http.StatusFound)
}

// GET /survey
// handleSurvey displays the survey if the user is logged in.
// If survey is already completed, user is redirected to the dashboard.
// If user is not logged in, user is redirected back to home.
func handleSurvey(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	if session.Id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ctx := appengine.NewContext(r)
	var user User
	key := datastore.NewKey(ctx, "User", session.Id, 0, nil)
	err := datastore.Get(ctx, key, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user.SurveyComplete {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	data := Data{
		Session:   session,
		Questions: questions,
		Answers:   answers,
	}
	serveTemplate(w, data, "layout", "navbar", "login", "register", "survey", "footer")
}

// POST /api/recordUserResponse
// recordUserResponse updates the user's survey responses and writes true.
// If the response was successfully recorded, false otherwise.
func recordUserResponse(w http.ResponseWriter, r *http.Request) {
	cUser, _ := r.Cookie("session-id")
	response := r.FormValue("response")
	username := cUser.Value
	if ok := updateUserResponses(w, r, username, response); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
}

// POST /api/aggregateResponses
// aggregateResponses retrieves the distribution of responses to each
// survey question in json format.
func aggregateResponses(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := datastore.NewQuery("User")
	var users []User
	_, err := u.GetAll(ctx, &users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var responses []int
	allResponses := [MAX_QUESTIONS][MAX_ANSWERS]int{}
	for i := 0; i < len(users); i++ {
		responses = users[i].Responses
		for j := 0; j < MAX_QUESTIONS; j++ {
			allResponses[j][responses[j]]++
		}
	}
	json.NewEncoder(w).Encode(allResponses)
}
