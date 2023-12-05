package entity

type Ticket struct {
	UserID  int
	EventID int
	Exp     int
	Status  bool
	Token   string
}
