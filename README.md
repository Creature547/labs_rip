# Лабораторная работа 4

Сервис отопительных приборов: Go, PostgreSQL, MinIO, Redis, Swagger.

После проверки логина и пароля сервер возвращает JWT. Передавайте его в заголовке `Authorization: Bearer <JWT>`. Защищённые методы проверяют подпись, срок действия и blacklist. При logout в Redis до истечения срока JWT сохраняется его идентификатор (jti), чтобы отозванный токен больше не принимался. Активные сессии в Redis не создаются.

Запуск: сохраните существующий `.env` в корне проекта, запустите PostgreSQL, MinIO и Redis, затем `go run ./cmd/awesomeProject`. Swagger: `http://localhost:8080/swagger/index.html`.

В Postman отключите автоматическую авторизацию и укажите Authorization вручную в Headers. Публичный GET работает без токена. PUT публикации собственного черновика требует действующий JWT. После POST /api/users/logout проверяйте Redis командой `docker exec heating_redis redis-cli --scan --pattern 'blacklist:*'`. Запись означает отозванный токен, не активную сессию.
