package parser

import "github.com/med-000/notifyclass/db"

type Course struct {
	ExternalId string
	Year       int
	Term       int
	Classes    []*Class
}

type Class struct {
	ExternalId string
	Day        int
	Period     int
	Title      string
	URL        string
	Events     []*Event
}

type Event struct {
	ExternalId string
	Name       string
	Category   string
	Date       string
	URL        string
	GroupName  string
	Content    []*Content
}

type Content struct {
	ContentType db.ContentType
	URL         string
	FileName    string
}
