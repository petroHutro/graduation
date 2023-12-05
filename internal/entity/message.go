package entity

type MessageTo struct {
	UserID int
	Mail   string
}

type Message struct {
	Users   []MessageTo
	EventID int
	Body    string
	Urls    []string
}
