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