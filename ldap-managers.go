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
	DN         string
	Name       string
	RusName    string
	Department string
	Title      string
	Manager    string
	RusManager string
	ManagerDN  string
}

// User structure
type User struct {
	Name string
	// Title string
	ManagedBy []string
	Managed   []string
}

func stringNotInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return false
		}
	}
	return true
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
		[]string{"mailNickname", "distinguishedName", "cn", "msDS-PhoneticDisplayName", "department", "title", "manager"},
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
				[]string{"mailNickname", "distinguishedName", "cn", "msDS-PhoneticDisplayName", "department", "title", "manager"},
				nil,
			)
			srr, err := l.Search(searchRequestManager)
			if err != nil {
				// logger.Info("No Object")
			} else {
				employee := Employee{
					Login:      entry.GetAttributeValue("mailNickname"),
					DN:         entry.GetAttributeValue("distinguishedName"),
					Name:       entry.GetAttributeValue("cn"),
					RusName:    entry.GetAttributeValue("msDS-PhoneticDisplayName"),
					Department: entry.GetAttributeValue("department"),
					Title:      entry.GetAttributeValue("title"),
					Manager:    srr.Entries[0].GetAttributeValue("cn"),
					RusManager: srr.Entries[0].GetAttributeValue("msDS-PhoneticDisplayName"),
					ManagerDN:  srr.Entries[0].GetAttributeValue("distinguishedName"),
				}
				employees = append(employees, employee)
			}
		}
	}

	var persons []User

	var managedUsers []string
	var managersUsers []string

	for _, employee := range employees {
		managedUsers = nil
		managersUsers = nil

		for _, managed := range employees {
			if managed.Manager == employee.Name && managed.Name != employee.Name {
				managedUsers = append(managedUsers, managed.Login)
			}
		}

		if employee.Name != employee.Manager {
			for _, manager := range employees {
				if manager.Name == employee.Manager && stringNotInSlice(manager.Login, managersUsers) {
					managersUsers = append(managersUsers, manager.Login)
					for _, manager2 := range employees {
						if manager2.Name == manager.Manager && stringNotInSlice(manager2.Login, managersUsers) {
							managersUsers = append(managersUsers, manager2.Login)
							for _, manager3 := range employees {
								if manager3.Name == manager2.Manager && stringNotInSlice(manager3.Login, managersUsers) {
									managersUsers = append(managersUsers, manager3.Login)
									for _, manager4 := range employees {
										if manager4.Name == manager3.Manager && stringNotInSlice(manager4.Login, managersUsers) {
											managersUsers = append(managersUsers, manager4.Login)
											for _, manager5 := range employees {
												if manager5.Name == manager4.Manager && stringNotInSlice(manager5.Login, managersUsers) {
													managersUsers = append(managersUsers, manager5.Login)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		person := User{employee.Login, managersUsers, managedUsers}
		persons = append(persons, person)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(persons)
	if err != nil {
		fmt.Println("Error")
	}
	json := string(jsonData)
	_ = ioutil.WriteFile("managers.json", []byte(json), 0644)
}
