package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"shuTodo/model"
	"shuTodo/service/token"
	"time"
)

type todoInput struct {
	Id           int64  `json:"id"`
	Content      string `json:"content"`
	Due          string `json:"due"`
	EstimateCost string `json:"estimate_cost"`
	Type         string `json:"type"`
}

func parseInput(input todoInput) (model.Todo, error) {
	due, err := time.Parse("2006-01-02T15:04:05.999999999Z", input.Due)
	if err != nil {
		return model.Todo{}, err
	}
	location, _ := time.LoadLocation("Asia/Shanghai")
	due = due.In(location)
	estimateCost, err := time.ParseDuration(input.EstimateCost)
	if err != nil {
		return model.Todo{}, err
	}
	return model.Todo{
		Id:           input.Id,
		Content:      input.Content,
		Due:          due,
		EstimateCost: estimateCost,
		Type:         input.Type,
	}, nil
}

func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenInHeader := r.Header.Get("Authorization")
	if len(tokenInHeader) <= 7 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	studentId := token.StudentIdForToken(tokenInHeader[7:])
	if studentId == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	var toCreateInput todoInput
	err = json.Unmarshal(body, &toCreateInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	toCreate, err := parseInput(toCreateInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	// todo: start a transaction?
	toCreate, err = model.SaveTodo(toCreate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = model.AssignTodoToStudent(studentId, toCreate.Id)
	// transaction end here
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenInHeader := r.Header.Get("Authorization")
	if len(tokenInHeader) <= 7 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	studentId := token.StudentIdForToken(tokenInHeader[7:])
	if studentId == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	todos, err := model.GetTodoByStudentId(studentId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	body, err := json.Marshal(todos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		GetTodoHandler(w, r)
	case "POST":
		CreateTodoHandler(w, r)
	case "PUT":
		CreateTodoHandler(w, r)
	}
}
