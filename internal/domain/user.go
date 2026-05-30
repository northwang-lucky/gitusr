package domain

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserFilter struct {
	Index *int
	Email *string
	Name  *string
}

type UserStore interface {
	List() ([]User, error)
	Add(user User) error
	Remove(index int) error
	SaveAll(users []User) error
	IsInitialized() bool
}
