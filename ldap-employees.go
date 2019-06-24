package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/ldap.v3"
)

// Employee structure
type Employee struct {
	Login      string
	ID         string
	DN         string
	Name       string
	RusName    string
	Department string
	Title      string
	Mobile     string
	Manager    string
	RusManager string
	ManagerDN  string
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		logger.Info("Error")
	}
	defer l.Close()

	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		logger.Info("Error")
	}
	err = l.Bind(login, password)
	if err != nil {
		logger.Info("Error")
	}

	searchRequest := ldap.NewSearchRequest(
		"OU",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalPerson)(manager=*))",
		[]string{"mailNickname", "employeeID", "distinguishedName", "cn", "msDS-PhoneticDisplayName", "department", "mobile", "title", "manager"},
		nil,
	)

	sr, err := l.SearchWithPaging(searchRequest, 20)
	if err != nil {
		logger.Info("Error")
	}

	var employees []Employee

	for _, entry := range sr.Entries {
		if strings.Contains(entry.GetAttributeValue("distinguishedName"), "OU=DES_Contact") ||
			strings.Contains(entry.GetAttributeValue("distinguishedName"), "OU=Disabled Users") ||
			strings.Contains(entry.GetAttributeValue("distinguishedName"), "test") {
		} else {
			searchRequestManager := ldap.NewSearchRequest(
				entry.GetAttributeValue("manager"),
				ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
				"(&(objectClass=organizationalPerson))",
				[]string{"cn", "msDS-PhoneticDisplayName", "distinguishedName"},
				nil,
			)
			srr, err := l.Search(searchRequestManager)
			if err != nil {
				// logger.Info("No Object")
			} else {
				employee := Employee{
					Login:      entry.GetAttributeValue("mailNickname"),
					ID:         entry.GetAttributeValue("employeeID"),
					DN:         entry.GetAttributeValue("distinguishedName"),
					Name:       entry.GetAttributeValue("cn"),
					RusName:    entry.GetAttributeValue("msDS-PhoneticDisplayName"),
					Department: entry.GetAttributeValue("department"),
					Mobile:     entry.GetAttributeValue("mobile"),
					Title:      entry.GetAttributeValue("title"),
					Manager:    srr.Entries[0].GetAttributeValue("cn"),
					RusManager: srr.Entries[0].GetAttributeValue("msDS-PhoneticDisplayName"),
					ManagerDN:  srr.Entries[0].GetAttributeValue("distinguishedName"),
				}
				employees = append(employees, employee)
			}
		}
	}

	var jsonData []byte
	jsonData, err = json.Marshal(employees)
	if err != nil {
		fmt.Println("Error")
	}
	json := string(jsonData)
	_ = ioutil.WriteFile("employees.json", []byte(json), 0644)
}
