package serializer

import "awesomeProject/internal/app/model"

type User struct {
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

func UserFromModel(user model.User) User {
	return User{ID: user.ID, Login: user.Login}
}
