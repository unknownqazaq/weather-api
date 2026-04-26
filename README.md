# Weather API (Go + chi)

HTTP-сервис для получения текущей погоды через общедоступные API:
- Open-Meteo Forecast API
- Open-Meteo Geocoding API
- CountriesNow API (список городов по стране)

---

## Запуск

```bash
go mod tidy
go run ./cmd/app
```

Сервер стартует на:

```
http://localhost:8080
```

---

## Проверка

### 🔹 Healthcheck

```bash
curl http://localhost:8080/health
```

Ответ:

```json
{"status":"ok"}
```

---

### 🔹 Получение погоды по координатам

```bash
curl "http://localhost:8080/api/weather?lat=43.2389&lon=76.8897"
```

Пример ответа:

```json
{
  "latitude": 43.2389,
  "longitude": 76.8897,
  "temperature": 18.4,
  "wind_speed": 7.2,
  "weather_code": 1,
  "time": "2026-04-14T14:00",
  "description": "Переменная облачность",
  "outfit_recommendation": "Куртка"
}
```

---

### 🔹 Погода по городу

```bash
curl "http://localhost:8080/weather/Almaty"
```

Пример ответа:

```json
{
  "city": "Almaty",
  "country": "Kazakhstan",
  "latitude": 43.25,
  "longitude": 76.95,
  "temperature": 18.4,
  "wind_speed": 7.2,
  "weather_code": 1,
  "time": "2026-04-14T14:00",
  "description": "Переменная облачность",
  "outfit_recommendation": "Куртка"
}
```

Рекомендации по температуре:
- холодно (`< 10°C`) — `Тёплая одежда`
- прохладно (`10°C - 19.9°C`) — `Куртка`
- тепло (`>= 20°C`) — `Лёгкая одежда`

---

### 🔹 Погода по городам страны

```bash
curl "http://localhost:8080/weather/country/Kazakhstan"
```

Возвращает погоду по 10 городам страны.

---

### 🔹 Топ-3 самых тёплых города страны

```bash
curl "http://localhost:8080/weather/country/Kazakhstan/top"
```

Возвращает только 3 города с самой высокой температурой (по убыванию).

---

## Live примеры ответов

Ниже ответы из реального запуска сервиса (значения температуры и времени меняются в зависимости от момента запроса).

### `/api/weather?lat=43.2389&lon=76.8897`

```json
{
  "latitude": 43.2389,
  "longitude": 76.8897,
  "temperature": 8.1,
  "wind_speed": 2.4,
  "weather_code": 61,
  "time": "2026-04-16T22:45",
  "description": "Дождь",
  "outfit_recommendation": "Тёплая одежда"
}
```

### `/weather/Almaty`

```json
{
  "city": "Almaty",
  "country": "Kazakhstan",
  "latitude": 43.25,
  "longitude": 76.91667,
  "temperature": 8.3,
  "wind_speed": 2.4,
  "weather_code": 61,
  "time": "2026-04-16T22:45",
  "description": "Дождь",
  "outfit_recommendation": "Тёплая одежда"
}
```

### `/weather/country/Kazakhstan/top`

```json
{
  "country": "Kazakhstan",
  "cities": [
    {
      "city": "Atyrau",
      "country": "Kazakhstan",
      "latitude": 47.11667,
      "longitude": 51.88333,
      "temperature": 9.1,
      "wind_speed": 16.6,
      "weather_code": 2,
      "time": "2026-04-16T22:45",
      "description": "Переменная облачность",
      "outfit_recommendation": "Тёплая одежда"
    },
    {
      "city": "Almaty",
      "country": "Kazakhstan",
      "latitude": 43.25,
      "longitude": 76.91667,
      "temperature": 8.3,
      "wind_speed": 2.4,
      "weather_code": 61,
      "time": "2026-04-16T22:45",
      "description": "Дождь",
      "outfit_recommendation": "Тёплая одежда"
    },
    {
      "city": "Baikonur",
      "country": "Kazakhstan",
      "latitude": 45.61667,
      "longitude": 63.31667,
      "temperature": 5.8,
      "wind_speed": 20.1,
      "weather_code": 2,
      "time": "2026-04-16T22:45",
      "description": "Переменная облачность",
      "outfit_recommendation": "Тёплая одежда"
    }
  ]
}
```

---

## Где взять координаты

Проще всего через Google Maps — клик по карте → копировать координаты.

---

## Структура

```
cmd/app          — точка входа
internal/handler — HTTP слой
internal/service — бизнес-логика
internal/client  — внешний API
```

---

## Стек

* Go
* net/http
* go-chi
* JSON
* Open-Meteo APIs
* CountriesNow API

---

Ran command: `git add .`
Ran command: `git commit -m "feat: add docker-compose and database schema for users and weather history"`
Ran command: `git push`

Супер, первый этап официально завершен и сохранен в истории! 🚀

Переходим к **Этапу 2: Слой данных и бизнес-логики**. Начинаем с фундамента — структуры `User`.

Отвечаю на свой же вопрос про `deleted_at`: так как в базе данных это поле может быть `NULL` (пустым), обычный тип `time.Time` нам не подойдет (Go выдаст ошибку при попытке записать туда NULL). Есть два варианта решения:
1. `sql.NullTime` — стандартный тип из пакета `database/sql`. Отлично работает с базой, но в JSON выглядит не очень красиво.
2. `*time.Time` (указатель на время) — **лучший вариант**. Указатель в Go может быть `nil`, что идеально соответствует SQL `NULL`. А при конвертации в JSON он тоже элегантно превратится в `null`.

**Ваша задача на сейчас:**
Создайте папку `internal/domain` и в ней файл `user.go`.
Заполните его структурой `User`:

```go
package domain

import "time"

type User struct {
	ID           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	FirstName    string     `db:"first_name" json:"first_name"`
	LastName     string     `db:"last_name" json:"last_name"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
```
*(`omitempty` означает, что если поле равно `nil`, оно вообще не попадет в JSON ответ)*

Как только создадите этот файл, дайте знать, и мы перейдем к самому "вкусному" — написанию слоя **Repository** (`internal/repository/postgres/user_repository.go`), где мы будем писать SQL-запросы на добавление и поиск пользователя!