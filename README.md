# Лабораторная работа № 4

**Тема:** аутентификация и разграничение доступа в сервисе отопительных приборов.

**Стек:** Go, Gin, GORM, PostgreSQL, Redis, MinIO, Swagger, Postman.

## Запуск

Запустить Docker Desktop, затем в терминале WSL:

```bash
docker start heating_postgres heating_redis minio_storage heating_adminer
cd ~/labs_rip/awesomeProject
go run ./cmd/awesomeProject
```

В корне проекта должен находиться `.env` с настройками PostgreSQL, Redis, MinIO и `JWT_SECRET`. Секреты в репозиторий не добавляются.

Swagger: `http://localhost:8080/swagger/index.html`. Если localhost недоступен, использовать IP WSL из команды `hostname -I` вместо localhost.

## Аутентификация и доступ

- `POST /api/users/auth` — вход по логину и паролю. Сервер возвращает JWT и устанавливает cookie `session_id`.
- В Redis хранится активная сессия: `session:<случайный идентификатор> → логин`, срок действия — 3600 секунд. Полный JWT в Redis не сохраняется.
- Защищённые запросы принимают `Cookie: session_id=...` либо `Authorization: Bearer ...`.
- Публичный список приборов доступен без входа. Публикация собственного черновика требует аутентификации.
- `POST /api/users/logout` удаляет текущую сессию. При выходе с JWT его идентификатор `jti` дополнительно помещается в blacklist до истечения срока действия.

## Проверка работы

1. Войти через Swagger в режиме инкогнито.
2. В Postman проверить `GET /api/heating-equipment` с авторизацией и без неё: список доступен в обоих случаях.
3. Проверить `PUT /api/heating-equipment/{id}/publish`: гостю — отказ 401/403, автору черновика — успешная публикация с обновлением `status` и `formed_at`.
4. До выхода показать активную сессию в Redis:

```bash
docker exec heating_redis redis-cli --scan --pattern 'session:*'
docker exec heating_redis redis-cli GET session:ID
docker exec heating_redis redis-cli TTL session:ID
```

В последних двух командах заменить `session:ID` полным ключом из первой команды. `GET` показывает логин пользователя, `TTL` — оставшееся время жизни сессии.

Для проверки гостя в Postman отключить автоматическую отправку cookies и убрать заголовки `Cookie` и `Authorization`.
