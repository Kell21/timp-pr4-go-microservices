# Task Manager — микросервисное веб-приложение на Go

Практическая работа №4 по дисциплине «Технологии и методы программирования».
Вариант 1 — **Task Manager** (добавление, удаление, список, изменение статуса задач).
Выполнил: Баулин А. В., группа БИСО-03-23.

## Архитектура

```
Браузер ──HTTP──▶ Web Service (Go, html/template) ──HTTP/JSON──▶ API Service (Go, REST) ──SQL──▶ PostgreSQL
```

* **API Service** (`api/`) — REST API `/api/tasks`, бизнес-логика (`internal/service`),
  паттерн Repository (`internal/repository`: PostgreSQL + in-memory), SQL-миграции (`migrations/`),
  применяемые автоматически при старте, повторные попытки подключения к БД.
* **Web Service** (`web/`) — HTML-страница со списком и формой, HTTP-клиент к API
  (`internal/client`), динамическое обновление статуса и удаление через `fetch`.
* **PostgreSQL 15** — таблица `tasks`.

Диаграммы (PlantUML): [`diagrams/component.puml`](diagrams/component.puml),
[`diagrams/sequence.puml`](diagrams/sequence.puml).

## REST API

| Метод | Путь | Описание | Ответ |
|-------|------|----------|-------|
| GET | `/health` | проверка работоспособности | 200 |
| GET | `/api/tasks` | список задач | 200 |
| POST | `/api/tasks` | `{"title": "..."}` — создать задачу | 201, 400, 422 |
| GET | `/api/tasks/{id}` | задача по id | 200, 400, 404 |
| PUT | `/api/tasks/{id}` | `{"title"?: "...", "done"?: true}` — изменить | 200, 400, 404, 422 |
| DELETE | `/api/tasks/{id}` | удалить задачу | 204, 400, 404 |

## Запуск

```bash
docker compose up --build
```

* веб-интерфейс — http://localhost:8080
* REST API — http://localhost:8081/api/tasks

Без Docker: запустить PostgreSQL, затем

```bash
cd api && DB_DSN="postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable" PORT=8081 go run ./cmd/server
cd web && API_URL=http://localhost:8081 PORT=8080 go run ./cmd/server
```

## Тесты

```bash
cd api && go test -v ./...      # + интеграционный тест, если задана TEST_DB_DSN
cd web && go test -v ./...
```

## CI/CD

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) (GitHub Actions):
**lint** (gofmt, golangci-lint c gosec) → **test** (go test, PostgreSQL как сервис-контейнер) →
**build-images** (многостадийные Dockerfile) → **scan** (trivy, HIGH/CRITICAL) и
**compose-smoke** (docker compose up + сценарий через веб-интерфейс) →
**push** (образы `ghcr.io/kell21/timp-pr4-go-microservices/{api,web}` с тегами `latest` и SHA).

Эквивалентный пайплайн для GitLab CI + GitLab Container Registry — [`.gitlab-ci.yml`](.gitlab-ci.yml).
