package models

type Topic struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type Video struct {
	Id    uint   `json:"id"`
	Name  string `json:"name"`
	Link  string `json:"link"`
	Topic Topic  `json:"topic"`
}
