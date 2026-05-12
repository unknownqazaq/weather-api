# Weather API

REST API сервис для управления пользователями, их городами и получения погоды.

## Стек

- Go, go-chi, sqlx, PostgreSQL
- golang-jwt/jwt/v5 (Аутентификация), bcrypt (Хеширование паролей)
- Open-Meteo API (погода + геокодинг)
- CountriesNow API (города по стране)
- pgx (драйвер PostgreSQL)

## Архитектура

```
cmd/app/                        — точка входа, graceful shutdown
internal/
├── auth/                       — генерация JWT и работа с контекстом
├── config/                     — конфигурация из env
├── client/                     — HTTP-клиент к внешним API
├── model/                      — структуры данных для БД (сущности)
├── dto/                        — структуры для API (Data Transfer Object)
├── middleware/                 — HTTP-middlewares (аутентификация, RBAC)
├── handler/                    — HTTP-хендлеры, роутер, хелперы
├── repository/postgres/        — слой работы с БД (sqlx)
└── service/                    — бизнес-логика и координация
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

### Auth (Аутентификация)

```bash
# Регистрация
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password_hash":"pass","first_name":"Ivan","last_name":"Ivanov"}'

# Логин (получение JWT)
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"pass"}'
```
В ответ придет `token`. Используйте его в заголовке `Authorization: Bearer <token>` для защищенных эндпоинтов.

### Users Profile

```bash
# Получить свой профиль
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <token>"
```

### Users Management (Admin Only)
Требуется токен администратора.

```bash
# Список пользователей
curl "http://localhost:8080/api/v1/users?limit=10&offset=0" \
  -H "Authorization: Bearer <admin_token>"

# Получить по ID
curl http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <admin_token>"

# Удалить (soft delete)
curl -X DELETE http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <admin_token>"
```

### User Cities (Protected)

```bash
# Добавить город (user_id берется из токена)
curl -X POST http://localhost:8080/api/v1/cities \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"city":"Almaty"}'

# Список городов пользователя
curl http://localhost:8080/api/v1/cities \
  -H "Authorization: Bearer <token>"

# Удалить город
curl -X DELETE http://localhost:8080/api/v1/cities/1 \
  -H "Authorization: Bearer <token>"
```

### Weather (Protected & Public)

```bash
# Погода по всем городам пользователя (параллельно) - Protected
curl http://localhost:8080/api/v1/weather \
  -H "Authorization: Bearer <token>"

# История погоды - Protected
curl "http://localhost:8080/api/v1/weather/history?city=Almaty&limit=10&offset=0" \
  -H "Authorization: Bearer <token>"

# Погода по координатам (Public)
curl "http://localhost:8080/api/weather_coords?lat=43.2389&lon=76.8897"

# Погода по городу (Public)
curl http://localhost:8080/weather/city/Almaty

# Погода по стране (топ-10 городов)
curl http://localhost:8080/weather/country/Kazakhstan

# Топ-3 самых теплых города страны
curl http://localhost:8080/weather/country/Kazakhstan/top
```

## Реализованные фичи

- **Clean Architecture**: строгое разделение на слои `Handler` → `Service` → `Repository`
- Разделение сущностей на `Model` (для БД) и `DTO` (для сокрытия приватных данных в ответах API)
- Выделение middleware в отдельный слой
- JWT аутентификация и авторизация (RBAC)
- Middleware для проверки токена и ролей (`user`, `admin`)
- Безопасное хеширование паролей (`bcrypt`)
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
| role | VARCHAR | Роль (`user` или `admin`) |
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