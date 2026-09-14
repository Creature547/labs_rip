# Лабораторная работа №2 — запуск

Проект переделан с in-memory коллекции на PostgreSQL + GORM. MinIO остается для изображений и видео.

## 1. Запустить хранилища

Из папки `awesomeProject`:

```bash
docker compose up -d
```

Проверка:

```bash
docker compose ps
```

Должны работать:
- PostgreSQL: `localhost:5432`
- Adminer: `http://localhost:8081`
- MinIO API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001`

## 2. Установить новые Go-зависимости

```bash
go mod tidy
```

Добавятся зависимости GORM, PostgreSQL driver и godotenv, а `go.sum` обновится автоматически.

## 3. Проверить .env

В проекте уже есть `.env`:

```dotenv
DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASS=mydbpass
DB_NAME=mydb
```

`.env` добавлен в `.gitignore`.

## 4. Создать таблицы миграцией

```bash
go run ./cmd/migrate
```

В БД создадутся ровно три предметные таблицы:
- `users`
- `heating_calculations`
- `heating_likes`

Каскадное удаление не используется. Дополнительно создается частичный уникальный индекс, запрещающий одному пользователю иметь больше одного расчета в статусе `черновик`.

## 5. Наполнить БД через Adminer

Открыть `http://localhost:8081`.

Параметры входа:
- System: PostgreSQL
- Server: `postgres`
- Username: `root`
- Password: `mydbpass`
- Database: `mydb`

После входа открыть `SQL command`, скопировать содержимое файла `seed_adminer.sql` и выполнить.

После выполнения будут:
- 3 опубликованных отопительных расчета;
- 1 удаленный расчет;
- 1 черновик пользователя с id=1;
- пользователи;
- лайки в таблице м-м.

Удаленный расчет в интерфейсе не показывается.

## 6. Запустить приложение

```bash
go run ./cmd/awesomeProject
```

Открыть:
- `http://localhost:8080/feed`
- `http://localhost:8080/draft`
- `http://localhost:8080/grid`

## Шесть HTTP-методов

### GET

1. `GET /feed?id=1` — получить опубликованный расчет по id; `?next=true` открывает следующий опубликованный.
2. `GET /draft` — получить черновик текущего пользователя (в лабораторной временно `currentUserID = 1`).
3. `GET /grid?min_power=5` — список опубликованных расчетов с серверной фильтрацией по мощности.

### POST

4. `POST /draft/create` — создать черновик через GORM. Поля: название, URL изображения, URL видео.
5. `POST /draft/publish` — заполнить описание, мощность и стоимость и опубликовать черновик через GORM.
6. `POST /heating/delete` — логическое удаление: статус меняется на `удален` сырым SQL `UPDATE` через `database/sql`, без ORM.

## Как проверить сценарий создания

Сейчас seed создает один черновик для пользователя 1, поэтому `/draft` сразу показывает второй шаг с кнопкой `Опубликовать`.

Чтобы проверить сценарий «черновика нет → кнопка Далее»:

1. В Adminer выполнить:

```sql
UPDATE heating_calculations
SET status = 'удален'
WHERE creator_id = 1 AND status = 'черновик';
```

2. Обновить `/draft`.
3. Заполнить название + URL картинки + URL видео и нажать `Далее`.
4. После редиректа заполнить описание, мощность и стоимость и нажать `Опубликовать`.
5. После публикации расчет появится в `/grid` и `/feed`.

## Что показывать преподавателю в коде

- `internal/app/model/heating.go` — три модели, связи и предметные поля.
- `internal/app/dsn/dsn.go` — DSN из переменных окружения.
- `cmd/migrate/main.go` — GORM AutoMigrate и ограничение одного черновика на пользователя.
- `internal/app/repository/heating_repository.go` — получение/поиск/создание/публикация через ORM и удаление через SQL.
- `internal/app/handler/heating_handler.go` — 6 обработчиков.
- `internal/api/server.go` — регистрация ровно трех GET и трех POST маршрутов.
- `docker-compose.yml` — PostgreSQL, Adminer, MinIO.
