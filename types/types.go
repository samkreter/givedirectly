package types


type Request struct {
	ID int `json:"id"`
	Email string `json:"email"`
	Title string `json:"title"`
}

type Book struct {
	ID int `json:"id"`
	Available bool `json:"available"`
	Title string `json:"title"`
	TimeRequested string `json:"timestamp"`
}
