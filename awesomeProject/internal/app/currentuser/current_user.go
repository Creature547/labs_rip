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


func Get() *User {
	once.Do(func() {
		instance = &User{ID: FixedUserID}
	})
	return instance
}
