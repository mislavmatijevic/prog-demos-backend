package models

type Topic struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type Video struct {
	Id         uint   `json:"id"`
	Name       string `json:"name"`
	Subtitle   string `json:"subtitle"`
	Identifier string `json:"identifier"`
	Topic      Topic  `json:"topic"`
}
