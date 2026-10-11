package currentuser

import "sync"

const FixedUserID uint = 1

type User struct {
	ID uint
}

var (
	once     sync.Once
	instance *User
)

// Get returns the single current-user instance used in lab 3.
// Authorization will replace this fixed user in lab 4.
func Get() *User {
	once.Do(func() {
		instance = &User{ID: FixedUserID}
	})
	return instance
}
