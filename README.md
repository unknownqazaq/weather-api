# Weather API (Go + chi)

HTTP-сервис для получения текущей погоды через внешний API (Open-Meteo).

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

### 🔹 Получение погоды

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
  "description": "Переменная облачность"
}
```

---

## Где взять координаты

Проще всего через Google Maps — клик по карте → копировать координаты.

---

## Структура

```
cmd/app         — точка входа
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
* Open-Meteo API

---# weather-api
