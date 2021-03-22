package types


type Request struct {
	ID int
	Email string
	Title string
}

type Book struct {
	ID int
	Available bool
	Title string
	TimeRequested string
}
