# In-Memory Key-Value Database with HTTP Interface

## Описание проекта

Этот проект представляет собой in-memory key-value базу данных с поддержкой HTTP интерфейса. Основная цель проекта — предоставить эффективное решение для хранения данных в оперативной памяти с возможностью персистентного хранения и автоматического удаления устаревших записей.

База данных поддерживает работу с различными типами данных, такими как скаляры, словари и массивы, а также реализует широкий набор операций для работы с этими типами.

---

## Основной функционал

- **Хранимые типы данных:**
  - Скаляры (строки и числа).
  - Словари.
  - Массивы.
- **Операции:**
  - Работа с ключами (`GET`, `SET`, `EXPIRE`).
  - Работа со словарями (`HSET`, `HGET`).
  - Работа с массивами (`LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LSET`, `LGET`).
- **Дополнительные возможности:**
  - Персистентное сохранение состояния на диск в формате JSON.
  - Автоматическое удаление устаревших записей (garbage collection).
  - Поддержка регулярных выражений для поиска ключей (`KEYS pattern`).
- **HTTP API:**
  - GET/POST запросы для взаимодействия с базой данных.
- **Docker и Docker Compose:**
  - Легкий запуск приложения и его базы данных PostgreSQL.

---

## Технологии

Проект был разработан с использованием следующих технологий и инструментов:

- **Язык программирования:** Golang (Concurrency, HTTP).
- **Фреймворк:** Gin для создания HTTP API.
- **Базы данных:**
  - In-memory database для быстрого доступа.
  - PostgreSQL для персистентного хранения.
  - JSON для сохранения состояния базы.
- **Контейнеризация:** Docker, Docker Compose.
- **Тестирование:** Встроенные возможности тестирования в Golang, бенчмаркинг.

---

## Установка и запуск

### Локально

1. Убедитесь, что у вас установлен Go (версия >= 1.18) и PostgreSQL.
2. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/chrizantona/go-course-2024.git
   cd go-course-2024
   ```
3. Установите зависимости и запустите приложение:
   ```bash
   go mod download
   go run cmd/main.go
   ```

### С использованием Docker Compose

1. Убедитесь, что у вас установлен Docker и Docker Compose.
2. Запустите контейнеры:
   ```bash
   docker-compose up --build
   ```
3. Приложение будет доступно по адресу: `http://localhost:8080`.

---

## HTTP API

### Примеры запросов

#### Работа со скалярами

**SET key value [EX seconds]**

```bash
curl -X POST http://localhost:8080/set -d 'key=name&value="Alan"&ex=20'
```

Ответ:

```json
{"status":"OK"}
```

**GET key**

```bash
curl -X GET http://localhost:8080/get?key=name
```

Ответ:

```json
"Alan"
```

#### Работа со словарями

**HSET key field value**

```bash
curl -X POST http://localhost:8080/hash/set -d 'key=user&field=name&value="Alan"'
```

Ответ:

```json
{"fields_affected":1}
```

**HGET key field**

```bash
curl -X GET http://localhost:8080/hash/get?key=user&field=name
```

Ответ:

```json
"Alan"
```

#### Работа с массивами

**LPUSH key element**

```bash
curl -X POST http://localhost:8080/array/lpush -d 'key=list&element=1'
```

Ответ:

```json
{"new_length":1}
```

**LPOP key**

```bash
curl -X POST http://localhost:8080/array/lpop -d 'key=list&count=1'
```

Ответ:

```json
{"elements":[1]}
```

---

## Архитектура

Проект организован следующим образом:

- **cmd/main.go:** Точка входа приложения.
- **internal/pkg:** Содержит основную бизнес-логику и модули приложения:
  - **server:** Реализация HTTP сервера и маршрутизации.
  - **storage:** Модуль для работы с in-memory базой данных и её персистентностью.
- **storage.json:** Файл для сохранения состояния базы данных.
- **Dockerfile:** Файл для контейнеризации приложения.
- **docker-compose.yml:** Конфигурация для запуска приложения и PostgreSQL через Docker Compose.

---

## Тестирование

Запуск тестов:

```bash
go test ./...
```

Запуск бенчмарков:

```bash
go test -bench=.
```

---

## Контакты

Если у вас есть вопросы или предложения, вы можете связаться со мной:

- Email: [sharukondondon@gmail.com](mailto\:sharukondondon@gmail.com)
- GitHub: [chrizantona](https://github.com/chrizantona)

---

## Лицензия

Этот проект лицензирован под MIT License. Подробнее в [LICENSE](./LICENSE).

