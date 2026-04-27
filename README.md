# Weather API

REST API сервис для управления пользователями, их городами и получения погоды.

## Стек

- Go, go-chi, sqlx, PostgreSQL
- Open-Meteo API (погода + геокодинг)
- CountriesNow API (города по стране)
- pgx (драйвер PostgreSQL)

## Архитектура

```
cmd/app/                        — точка входа, graceful shutdown
internal/
├── config/                     — конфигурация из env
├── client/                     — HTTP-клиент к внешним API
├── domain/                     — доменные структуры и валидация
├── handler/                    — HTTP-хендлеры, роутер, хелперы
├── repository/postgres/        — слой работы с БД (sqlx)
└── service/                    — бизнес-логика
```

## Запуск

```bash
docker-compose up -d
go mod tidy
go run ./cmd/app
```

Сервер стартует на `http://localhost:8080`

## База данных

PostgreSQL запускается через Docker на порту `5433`:

| Параметр | Значение   |
|----------|------------|
| Host     | localhost  |
| Port     | 5433       |
| Database | users_db   |
| User     | postgres   |
| Password | postgres   |

## API Endpoints

### Healthcheck

```bash
curl http://localhost:8080/health
```

### Users CRUD

```bash
# Создать пользователя
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password_hash":"pass","first_name":"Ivan","last_name":"Ivanov"}'

# Список пользователей (пагинация + поиск)
curl "http://localhost:8080/api/v1/users?limit=10&offset=0&q=Ivan"

# Получить по ID
curl http://localhost:8080/api/v1/users/1

# Обновить
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Petr"}'

# Удалить (soft delete)
curl -X DELETE http://localhost:8080/api/v1/users/1
```

### User Cities

```bash
# Добавить город
curl -X POST http://localhost:8080/api/v1/users/1/cities \
  -H "Content-Type: application/json" \
  -d '{"city":"Almaty"}'

# Список городов пользователя
curl http://localhost:8080/api/v1/users/1/cities

# Удалить город
curl -X DELETE http://localhost:8080/api/v1/users/1/cities/1
```

### Weather

```bash
# Погода по всем городам пользователя (параллельные запросы + кеш 1 мин)
curl http://localhost:8080/api/v1/users/1/weather

# История погоды (фильтр по городу, пагинация)
curl "http://localhost:8080/api/v1/users/1/weather/history?city=Almaty&limit=10&offset=0"

# Погода по координатам
curl "http://localhost:8080/api/weather?lat=43.2389&lon=76.8897"

# Погода по городу
curl http://localhost:8080/weather/Almaty

# Погода по стране (топ-10 городов)
curl http://localhost:8080/weather/country/Kazakhstan

# Топ-3 самых теплых города страны
curl http://localhost:8080/weather/country/Kazakhstan/top
```

## Реализованные фичи

- CRUD пользователей с soft delete
- Управление городами пользователя (добавление, список, удаление)
- Параллельный запрос погоды по всем городам пользователя (goroutines + sync.WaitGroup)
- Автоматическое сохранение истории запросов погоды в БД
- Фильтрация истории по городу с пагинацией (limit/offset)
- Поиск пользователей по имени/email с пагинацией
- Нормализация входных параметров (Normalize)
- Graceful Shutdown (корректная остановка сервера)
- Connection Pool для PostgreSQL (pgx)
- Защита от SQL-инъекций (sqlx.Named + Rebind)
- Динамическое построение SQL через strings.Builder
- Вынос роутера и хелперов в отдельные файлы
- Строгие типизированные JSON-ответы (без map[string]interface{})

## Структура базы данных

### Таблица `users`

| Поле | Тип | Описание |
|------|-----|----------|
| id | SERIAL PRIMARY KEY | ID пользователя |
| email | VARCHAR UNIQUE | Email (уникальный) |
| password_hash | VARCHAR | Хеш пароля (bcrypt) |
| first_name | VARCHAR | Имя |
| last_name | VARCHAR | Фамилия |
| created_at | TIMESTAMP | Дата создания |
| deleted_at | TIMESTAMP NULL | Дата удаления (soft delete) |

### Таблица `user_cities`

| Поле | Тип | Описание |
|------|-----|----------|
| id | SERIAL PRIMARY KEY | ID записи |
| user_id | INT REFERENCES users | ID пользователя |
| city | VARCHAR | Название города |
| added_at | TIMESTAMP | Дата добавления |

### Таблица `weather_history`

| Поле | Тип | Описание |
|------|-----|----------|
| id | SERIAL PRIMARY KEY | ID записи |
| user_id | INT REFERENCES users | ID пользователя |
| city | VARCHAR | Город |
| temperature | DECIMAL | Температура (°C) |
| description | VARCHAR | Описание погоды |
| requested_at | TIMESTAMP | Время запроса |

### Пример данных

```
users:
 id |        email        | first_name | last_name | deleted_at
----+---------------------+------------+-----------+------------
  1 | test@test.com       | TestNamed  | Ivanov    |
  2 | test2@test.com      | Anna-Maria | Karenina  | (удалён)
  6 | pgx_test@test.com   | PGX        | Test      |

user_cities:
 id | user_id |  city
----+---------+--------
  1 |       1 | Almaty
  2 |       2 | Astana
  5 |       6 | Tokyo

weather_history:
 id | user_id |  city  | temperature |      description
----+---------+--------+-------------+-----------------------
  1 |       1 | Almaty |       12.40 | Переменная облачность
  3 |       2 | Astana |        3.00 | Ясно
  7 |       6 | Tokyo  |       13.70 | Дождь
```