# Bank REST API


## 📋 Описание

Безопасный REST API для банковского сервиса. Реализованы регистрация и аутентификация пользователей, управление счетами и картами, кредитные операции, переводы средств и финансовая аналитика.

### Основные возможности

-  **Регистрация и аутентификация** с JWT-токенами
- **Управление счетами** (создание, пополнение, снятие, переводы)
- **Виртуальные карты** (генерация по алгоритму Луна, шифрование PGP)
-  **Кредиты** (оформление, аннуитетные платежи, автоматическое списание)
- **Финансовая аналитика** (доходы/расходы, кредитная нагрузка, прогноз баланса)
-  **Email-уведомления** (SMTP)
-  **Интеграция с ЦБ РФ** (получение ключевой ставки через SOAP)

## 🛠 Технологический стек

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.23+ |
| Маршрутизация | Gorilla Mux |
| База данных | PostgreSQL 17 + pgcrypto |
| Аутентификация | JWT (golang-jwt/jwt/v5) |
| Логирование | Logrus |
| Шифрование | bcrypt, HMAC-SHA256, PGP |
| Парсинг XML | Beevik/etree |
| Email | Gomail v2 |
| Контейнеризация | Docker, Docker Compose |

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.23+
- Docker и Docker Compose
- PGP ключи (для шифрования карт)

### Установка и запуск

1. **Клонируйте репозиторий**

```bash
git clone https://github.com/Dimawalker/bank.git
cd bank

Настройте окружение

bash
cp .env.example .env
# Отредактируйте .env, добавьте JWT_SECRET и другие переменные
Запустите базу данных через Docker

bash
docker-compose up -d
Установите зависимости

bash
go mod download
Запустите приложение

bash
go run cmd/main.go



Примеры запросов
Регистрация пользователя
bash
curl -X POST http://localhost:8080/api/v1/public/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
Аутентификация
bash
curl -X POST http://localhost:8080/api/v1/public/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
Создание счёта (с JWT)
bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "currency": "RUB"
  }'
Выпуск карты
bash
curl -X POST http://localhost:8080/api/v1/cards \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "card_type": "debit"
  }'
Оформление кредита
bash
curl -X POST http://localhost:8080/api/v1/credits \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100000,
    "term_months": 12,
    "interest_rate": 15.5
  }'
