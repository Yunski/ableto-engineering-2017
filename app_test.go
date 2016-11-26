package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

var inst aetest.Instance

func TestMain(m *testing.M) {
	var err error
	inst, err = aetest.NewInstance(nil)
	if err != nil {
		log.Fatalf("failed to create new Instance: %v", err)
	}
	defer inst.Close()
	e := m.Run()
	os.Exit(e)
}

func TestCreateUser(t *testing.T) {
	username := "User"
	password := "password"
	params := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ := inst.NewRequest("POST", "/createusers", strings.NewReader(params))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w := httptest.NewRecorder()
	// create new user
	createUser(w, r)
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", username, 0, nil)
	var user User
	err := datastore.Get(ctx, key, &user)
	if user.Id != username {
		t.Error("Expected user to have username:", username)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		t.Error("Expected user to have password: ", password)
	}
	// create user that already exists
	password = "changed"
	params = fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ = inst.NewRequest("POST", "/createusers", strings.NewReader(params))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	createUser(w, r)
	err = datastore.Get(ctx, key, &user)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err == nil {
		t.Error("Did not expect password to be changed")
	}
	datastore.Delete(ctx, key)
	datastore.Get(ctx, key, &user)
}

func TestLoginAndLogout(t *testing.T) {
	username := "User"
	password := "password"
	params := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ := inst.NewRequest("POST", "/createusers", strings.NewReader(params))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w := httptest.NewRecorder()
	createUser(w, r)
	//login success
	paramsSuccess := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ = inst.NewRequest("POST", "/login", strings.NewReader(paramsSuccess))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	login(w, r)
	content, _ := ioutil.ReadAll(w.Body)
	if string(content) != "true" {
		t.Error("Expected login to succeed for user")
	}
	//login failure
	paramsFailure := fmt.Sprintf("username=%s&password=fail", username)
	r, _ = inst.NewRequest("POST", "/login", strings.NewReader(paramsFailure))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	login(w, r)
	content, _ = ioutil.ReadAll(w.Body)
	if string(content) == "true" {
		t.Error("Expected login to fail for user")
	}
	// login nonexistent user
	paramsNonExist := fmt.Sprintf("username=nonuser&password=fail")
	r, _ = inst.NewRequest("POST", "/login", strings.NewReader(paramsNonExist))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	login(w, r)
	content, _ = ioutil.ReadAll(w.Body)
	if string(content) == "true" {
		t.Error("Expected login to fail for nonexistent user")
	}
	// test user who has completed survey
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := User{
		Id:             username,
		Password:       string(hash[:]),
		Responses:      []int{},
		SurveyComplete: true,
	}
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", user.Id, 0, nil)
	datastore.Put(ctx, key, &user)
	paramsComplete := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ = inst.NewRequest("POST", "/login", strings.NewReader(paramsComplete))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	login(w, r)
	content, _ = ioutil.ReadAll(w.Body)
	if string(content) != "true" {
		t.Error("Expected login to succeed for user")
	}
	r, _ = inst.NewRequest("POST", "/logout", nil)
	w = httptest.NewRecorder()
	logout(w, r)
	if w.Code != http.StatusFound {
		t.Error("Logout failed")
	}
	datastore.Delete(ctx, key)
	datastore.Get(ctx, key, &user)
}

func TestHome(t *testing.T) {
	r, _ := inst.NewRequest("GET", "/home", nil)
	w := httptest.NewRecorder()
	home(w, r)
	if w.Code != http.StatusOK {
		t.Error("Failed to access home page")
	}
	// access home with account
	username := "User"
	password := "password"
	params := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ = inst.NewRequest("POST", "/createusers", strings.NewReader(params))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	createUser(w, r)
	r, _ = inst.NewRequest("GET", "/home", nil)
	addCookies(r, username)
	w = httptest.NewRecorder()
	home(w, r)
	if w.Code != http.StatusOK {
		t.Error("Failed to access home page")
	}
	// access home when completed survey
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := User{
		Id:             username,
		Password:       string(hash[:]),
		Responses:      []int{},
		SurveyComplete: true,
	}
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", user.Id, 0, nil)
	datastore.Put(ctx, key, &user)
	r, _ = inst.NewRequest("GET", "/home", nil)
	addCookies(r, user.Id)
	w = httptest.NewRecorder()
	home(w, r)
	if w.Code != http.StatusFound {
		t.Error("Failed to redirect to dashboard")
	}
	datastore.Delete(ctx, key)
	datastore.Get(ctx, key, &user)
}

func TestAbout(t *testing.T) {
	r, _ := inst.NewRequest("GET", "/about", nil)
	w := httptest.NewRecorder()
	about(w, r)
	if w.Code != http.StatusOK {
		t.Error("Failed to get about page")
	}
}

func TestSurvey(t *testing.T) {
	// attempt to take survey without account
	r, _ := inst.NewRequest("POST", "/survey", nil)
	w := httptest.NewRecorder()
	handleSurvey(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Error("Unexpected access to survey without logging in")
	}
	username := "User"
	password := "password"
	params := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ = inst.NewRequest("POST", "/createusers", strings.NewReader(params))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	createUser(w, r)
	// take survey with existing account
	r, _ = inst.NewRequest("POST", "/survey", nil)
	addCookies(r, username)
	w = httptest.NewRecorder()
	handleSurvey(w, r)
	if w.Code != http.StatusOK {
		t.Error("Failed to take survey")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := User{
		Id:             username,
		Password:       string(hash[:]),
		Responses:      []int{},
		SurveyComplete: true,
	}
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", user.Id, 0, nil)
	datastore.Put(ctx, key, &user)
	// attempt to take survey when already complete
	r, _ = inst.NewRequest("POST", "/survey", nil)
	addCookies(r, user.Id)
	w = httptest.NewRecorder()
	handleSurvey(w, r)
	if w.Code != http.StatusFound {
		t.Error("Failed to redirect to dashboard")
	}
	datastore.Delete(ctx, key)
	datastore.Get(ctx, key, &user)
}

func TestDashboard(t *testing.T) {
	// get dashboard without logging in
	r, _ := inst.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	dashboard(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Error("Unexpected access to dashboard")
	}
	username := "User"
	password := "password"
	params := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ = inst.NewRequest("POST", "/createusers", strings.NewReader(params))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w = httptest.NewRecorder()
	createUser(w, r)
	// get dashboard without completing survey
	r, _ = inst.NewRequest("GET", "/dashboard", nil)
	addCookies(r, username)
	w = httptest.NewRecorder()
	dashboard(w, r)
	if w.Code == http.StatusOK {
		t.Error("Unexpected access to dashboard")
	}
	// get dashboard after completing survey
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := User{
		Id:             username,
		Password:       string(hash[:]),
		Responses:      []int{1, 1, 1, 1},
		SurveyComplete: true,
	}
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", user.Id, 0, nil)
	datastore.Put(ctx, key, &user)
	r, _ = inst.NewRequest("GET", "/dashboard", nil)
	addCookies(r, user.Id)
	w = httptest.NewRecorder()
	dashboard(w, r)
	if w.Code != http.StatusOK {
		t.Error("failed to access dashboard")
	}
	datastore.Delete(ctx, key)
	datastore.Get(ctx, key, &user)
}

func TestRecordUserResponse(t *testing.T) {
	username := "User"
	password := "password"
	params := fmt.Sprintf("username=%s&password=%s", username, password)
	r, _ := inst.NewRequest("POST", "/createusers", strings.NewReader(params))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w := httptest.NewRecorder()
	createUser(w, r)
	testResponse := "1"
	responseParam := fmt.Sprintf("response=%s", testResponse)
	// test record with invalid session
	cUser := &http.Cookie{
		Name:  "session-id",
		Value: "-1",
	}
	r, _ = inst.NewRequest("POST", "/api/recordUserResponse", strings.NewReader(responseParam))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	r.AddCookie(cUser)
	w = httptest.NewRecorder()
	recordUserResponse(w, r)
	if w.Code == http.StatusOK {
		t.Fatal("Unexpected success with invalid session")
	}
	// test record with valid session
	r, _ = inst.NewRequest("POST", "/api/recordUserResponse", strings.NewReader(responseParam))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	addCookies(r, username)
	w = httptest.NewRecorder()
	recordUserResponse(w, r)
	if w.Code != http.StatusOK {
		t.Fatal("Failed to record user response")
	}
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "User", username, 0, nil)
	var user User
	datastore.Get(ctx, key, &user)
	if len(user.Responses) != 1 {
		t.Error("failed to record user response")
	}
	if user.Responses[0] != 1 {
		t.Error("incorrect recorded response")
	}
	// test finish survey
	for i := 1; i < MAX_ANSWERS; i++ {
		r, _ = inst.NewRequest("POST", "/api/recordUserResponse", strings.NewReader(responseParam))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		addCookies(r, username)
		w = httptest.NewRecorder()
		recordUserResponse(w, r)
		if w.Code != http.StatusOK {
			t.Fatal("Failed to record user response")
		}
	}
	datastore.Get(ctx, key, &user)
	if len(user.Responses) != 4 {
		t.Error("failed to record user response")
	}
	if !user.SurveyComplete {
		t.Error("failed to update survey completion")
	}
	for i := 0; i < MAX_ANSWERS; i++ {
		if user.Responses[0] != 1 {
			t.Error("incorrect recorded response")
			break
		}
	}
	datastore.Delete(ctx, key)
	datastore.Get(ctx, key, &user)
}

func TestGetUserResponses(t *testing.T) {
	username1 := "User1"
	username2 := "User2"
	password := "password"
	user1 := User{
		Id:             username1,
		Password:       password,
		Responses:      []int{0, 0, 0, 0},
		SurveyComplete: true,
	}
	user2 := User{
		Id:             username2,
		Password:       password,
		Responses:      []int{1, 1, 1, 1},
		SurveyComplete: true,
	}
	r, _ := inst.NewRequest("POST", "/api/aggregateResponses", nil)
	ctx := appengine.NewContext(r)
	key1 := datastore.NewKey(ctx, "User", username1, 0, nil)
	datastore.Put(ctx, key1, &user1)
	datastore.Get(ctx, key1, &user1)
	key2 := datastore.NewKey(ctx, "User", username2, 0, nil)
	datastore.Put(ctx, key2, &user2)
	datastore.Get(ctx, key2, &user2)
	u := datastore.NewQuery("User")
	var users []User
	u.GetAll(ctx, &users)
	w := httptest.NewRecorder()
	aggregateResponses(w, r)
	expected := [][]int{
		{1, 1, 0, 0},
		{1, 1, 0, 0},
		{1, 1, 0, 0},
		{1, 1, 0, 0},
	}
	output := [][]int{}
	json.NewDecoder(w.Body).Decode(&output)
	if len(output) != MAX_QUESTIONS {
		t.Fatal("error decoding response")
	}
	for i := 0; i < MAX_QUESTIONS; i++ {
		for j := 0; j < MAX_ANSWERS; j++ {
			if expected[i][j] != output[i][j] {
				t.Error("incorrect aggregate response")
			}
		}
	}
	datastore.Delete(ctx, key1)
	datastore.Delete(ctx, key2)
	datastore.Get(ctx, key1, &user1)
	datastore.Get(ctx, key2, &user2)
}

func addCookies(r *http.Request, id string) {
	cUser := &http.Cookie{
		Name:  "session-id",
		Value: id,
	}
	cQuestion := &http.Cookie{
		Name:  "current-question",
		Value: strconv.Itoa(1),
	}
	qIndex := &http.Cookie{
		Name:  "current-qindex",
		Value: strconv.Itoa(0),
	}
	r.AddCookie(cUser)
	r.AddCookie(cQuestion)
	r.AddCookie(qIndex)
}
