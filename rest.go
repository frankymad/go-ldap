package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Employee struct {
	Login      string `json:"Login"`
	ID         string `json:"ID"`
	DN         string `json:"DN"`
	Name       string `json:"Name"`
	RusName    string `json:"RusName"`
	Department string `json:"Department"`
	Title      string `json:"Title"`
	Mobile     string `json:"Mobile"`
	Manager    string `json:"Manager"`
	RusManager string `json:"RusManager"`
	ManagerDN  string `json:"ManagerDN"`
}

type UserManagers struct {
	Name      string   `json:"Name"`
	ManagedBy []string `json:"ManagedBy"`
	Managed   []string `json:"Managed"`
}

var employees []Employee
var managers []UserManagers

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/user/{login}", getUser).Methods("GET")
	r.HandleFunc("/userManager/{name}", getManagers).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func getUser(w http.ResponseWriter, r *http.Request) {
	employeesFile, _ := ioutil.ReadFile("employees.json")
	json.Unmarshal(employeesFile, &employees)
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, user := range employees {
		if strings.ToLower(user.Login) == params["login"] {
			json.NewEncoder(w).Encode(user)
			return
		}
	}
	json.NewEncoder(w).Encode(&Employee{})
}

func getManagers(w http.ResponseWriter, r *http.Request) {
	managersFile, _ := ioutil.ReadFile("managers.json")
	json.Unmarshal(managersFile, &managers)
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, user := range managers {
		if strings.ToLower(user.Name) == strings.ToLower(params["name"]) {
			json.NewEncoder(w).Encode(user)
			return
		}
	}
	json.NewEncoder(w).Encode(&UserManagers{})
}
