package main

type Topic struct {
	User  string
	Error string

	Topic     string `json:"topic"`
	Filename  string `json:"filename"`
	Query1    string `json:"query1"`
	Query2    string `json:"query2"`
	Completed bool   `json:"completed"`
}
