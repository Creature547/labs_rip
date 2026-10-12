package main

import "awesomeProject/internal/api"

// @title Отопительные приборы — ЛР4
// @version 4.0
// @description JWT + blacklist Redis. В Authorize вводите Bearer и токен через пробел.
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	api.StartServer()
}
