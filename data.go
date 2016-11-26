package main

// User model
type User struct {
	Id             string
	Password       string
	Responses      []int
	SurveyComplete bool
}

// Session model
type Session struct {
	User
	LoggedIn      bool
	QuestionIndex int
	CurQuestion   int
}

// Data model for templates
type Data struct {
	Session   Session
	Questions []string
	Answers   [][]string
	Responses []string
}
