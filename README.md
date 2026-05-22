# Organization API

REST API для управления организационной структурой: подразделения и сотрудники в виде дерева.

## Стек

- Язык: **Go** (net/http)
- БД: **PostgreSQL** + **GORM**
- Миграции: **goose**
- Контейниризация: **Docker** + **docker-compose**
- Логирование: **log/slog**
- Тестирование: **testify** + **httptest**

## Запуск

```bash
cp .env.example .env
docker-compose up --build
```

Сервис доступен на `http://localhost:8080`. Миграции применяются автоматически при старте.

## Переменные окружения

- `DB_HOST` - Хост PostgreSQL
- `DB_PORT` - Порт PostgreSQL
- `DB_USER` - Пользователь БД
- `DB_PASSWORD` - Пароль БД
- `DB_NAME` - Имя БД
- `DB_SSLMODE` - SSL-режим
- `DB_MAX_OPEN_CONNS` - Макс. открытых соединений
- `DB_MAX_IDLE_CONNS` - Макс. простаивающих соединений
- `DB_CONN_MAX_LIFETIME` - Время жизни соединения
- `HTTP_ADDR` - Адрес HTTP-сервера
- `HTTP_READ_TIMEOUT` - Таймаут чтения запроса
- `HTTP_WRITE_TIMEOUT` - Таймаут записи ответа
- `HTTP_IDLE_TIMEOUT` - Keep-alive таймаут
- `LOG_LEVEL` - Уровень логирования (`debug`, `info`, `warn`, `error`)

## API

### Подразделения

#### Создать подразделение
```
POST /departments/
```
```json
{ "name": "Engineering", "parent_id": 1 }
```

#### Получить подразделение с деревом
```
GET /departments/{id}?depth=1&include_employees=true
```
- `depth` - глубина дерева, от 1 до 5 (по умолчанию `1`)
- `include_employees` - включить сотрудников (`true` / `false`)

#### Обновить подразделение
```
PATCH /departments/{id}
```
```json
{ "name": "New Name", "parent_id": 2 }
```

#### Удалить подразделение
```
DELETE /departments/{id}?mode=cascade
DELETE /departments/{id}?mode=reassign&reassign_to_department_id=5
```
- `cascade` - удалить подразделение вместе со всеми дочерними и сотрудниками
- `reassign` - перепривязать дочерние подразделения и сотрудников к `reassign_to_department_id`, затем удалить

### Сотрудники

#### Создать сотрудника
```
POST /departments/{id}/employees/
```
```json
{ "full_name": "Иван Иванов", "position": "Developer", "hired_at": "2024-01-15" }
```

### Коды ошибок

400 -`INVALID_INPUT` Невалидный запрос
404 -`NOT_FOUND` Ресурс не найден
409 - `ALREADY_EXISTS` Подразделение с таким именем уже существует
500 - `INTERNAL_ERROR` Внутренняя ошибка сервера

Формат ошибки:
```json
{ "code": "NOT_FOUND", "message": "department not found" }
```

## Тесты

```bash
go test ./...
```

45 тестов: юнит-тесты сервисного слоя (20) и тесты хендлеров через httptest (25).

## Принятые решения

### Архитектура

Проект разделён на четыре слоя: `transport -> service -> repository -> domain`. Каждый слой общается с соседним только через интерфейсы - это позволяет тестировать каждый слой изолированно с моками.

### Обработка ошибок

Введён пакет `apperr` с sentinel-ошибками (`ErrNotFound`, `ErrAlreadyExists`) и типом `apperr.Error{Code, Message}`. Репозиторий оборачивает DB-ошибки в sentinels, сервис конвертирует их в `apperr.Error` с понятным сообщением, хендлер маппит код на HTTP-статус. Такой подход избегает циклических импортов между слоями.

### Транзакции (TxRunner)

Для атомарного удаления с reassign нужна транзакция, которая затрагивает два репозитория. Вместо того чтобы прокидывать `*gorm.DB` в сервис, введён интерфейс `TxRunner.RunInTx(fn func(DeptRepo, EmpRepo) error)`. Реализация в пакете `repository` создаёт tx-scoped копии репозиториев через `WithTx(tx)` и передаёт их в коллбек. Сервис остаётся чистым от деталей GORM.

### Обнаружение циклов

`PATCH /departments/{id}` может переместить подразделение в другое место дерева. Чтобы не допустить цикл (dept A -> dept B -> dept A), перед обновлением выполняется проверка через рекурсивный CTE:

```sql
WITH RECURSIVE tree AS (
    SELECT id FROM departments WHERE id = $1
    UNION ALL
    SELECT d.id FROM departments d JOIN tree t ON d.parent_id = t.id
)
SELECT COUNT(*) FROM tree WHERE id = $2
```

Если новый родитель находится в поддереве текущего департамента - возвращается 400.

### DELETE reassign

По ТЗ reassign переносит сотрудников в другой департамент. В реализации условие расширено: дочерние подразделения удаляемого тоже перепривязываются к `reassign_to`, чтобы сохранить их вместе с их поддеревьями. Три операции выполняются в одной транзакции: `ReassignAll` (сотрудники) -> `ReparentChildren` (прямые дети) -> `Delete`. Перед транзакцией проверяется, что `reassign_to` не находится внутри поддерева удаляемого (проверка через тот же CTE, внутри транзакции).

### Миграции

Файлы миграций вшиты в бинарник через `//go:embed` и запускаются автоматически при старте через `goose.Up()`. Не требуется монтирование файлов в Docker и отдельный entrypoint-скрипт.

### Ответы API

Доменные модели (`domain.Department`, `domain.Employee`, `domain.DepartmentTree`) используются напрямую в качестве ответов - без отдельного слоя response-DTO. JSON-теги заданы на доменных структурах. Это уменьшает количество алокаций при сохранении гибкости.
