package entity

type Message struct {
	UserID  int
	EventID int
	Mail    string
	Body    string
	Urls    []string
}
