# Mercury Backend

B2B/B2C E-commerce Platform на микросервисах.

---

## О проекте

Mercury — это платформа для электронной коммерции с микросервисной архитектурой.

**Основные возможности:**

- Каталог товаров (products)
- Управление заказами (orders)
- Обработка платежей (payments)
- Асинхронная обработка через Kafka
- Real-time уведомления через WebSocket

**Технологии:**

- Go 1.25+
- PostgreSQL 15
- Redis 7
- Apache Kafka
- gRPC + REST Gateway
- Clean Architecture

---

## Структура проекта

mercury-backend/
├── cmd/ # Точки входа сервисов
├── internal/ # Приватный код (Clean Architecture)
├── pkg/ # Общие библиотеки
├── api/proto/ # Git submodule (proto-контракты)
├── configs/ # Конфигурационные файлы
├── migrations/sql/ # SQL миграции
├── docker/ # Dockerfile для сервисов
├── tests/ # E2E и нагрузочные тесты
├── docker-compose.yml # Инфраструктура
├── Makefile # Команды разработки
├── .env # Переменные окружения
└── README.md

---

## Быстрый старт

### Требования

- Go 1.25+
- Docker & Docker Compose
- Make
- golang-migrate

### Установка

```bash
# 1. Клонировать репозиторий
git clone https://github.com/your-username/mercury-backend.git
cd mercury-backend

# 2. Инициализировать submodule (proto)
git submodule update --init --recursive

# 3. Настроить окружение
cp .env.example .env

# 4. Запустить инфраструктуру
make docker-up

# 5. Применить миграции
make migrate-up

# 6. Собрать сервисы
make build

# 7. Запустить сервисы
make run
```

#Проверка

## Проверить статус инфраструктуры

docker-compose ps

## Проверить миграции

make migrate-version

## Подключиться к БД

make db-shell

## Запустить тесты

make test

# Команды разработки

## Инфраструктура

make docker-up # Запустить
make docker-down # Остановить
make docker-clean # Очистить всё

## Миграции

make migrate-up # Применить
make migrate-down # Откатить
make migrate-version # Проверить версию
make db-shell # Подключиться к БД

## Сборка

make build # Собрать сервисы
make run # Запустить локально
make clean # Очистить

## Тесты

make test # Unit-тесты
make test-coverage # С покрытием

## Качество кода

make fmt # Форматировать
make lint # Линтер

#Контакты

GitHub: https://github.com/FrishStrike/mercury-backend
