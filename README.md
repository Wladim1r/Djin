# 📊 Dzhinn Statistics Counter

## Описание проекта

**Dzhinn Statistics Counter** - это веб-приложение для учета и мониторинга ежедневной статистики по различным показателям в региональном разрезе. Система позволяет пользователям вводить, редактировать и просматривать статистические данные с автоматическим агрегированием по регионам.

## 🏗️ Архитектура

Проект построен по принципу **Clean Architecture** с разделением на слои:

```
internal/
├── api/
│   ├── handler/     # HTTP обработчики (контроллеры)
│   ├── service/     # Бизнес-логика
│   └── repository/  # Работа с базой данных
├── auth/           # Система авторизации и аутентификации
├── db/             # Конфигурация базы данных
├── lib/            # Общие библиотеки и утилиты
└── models/         # Структуры данных
```

## 🛠️ Технологический стек

- **Backend**: Go 1.x
- **Web Framework**: Gin
- **ORM**: GORM
- **База данных**: PostgreSQL
- **Сессии**: Cookie-based sessions
- **Авторизация**: bcrypt для хеширования паролей
- **Frontend**: HTML templates, статические файлы

## 📈 Основной функционал

### Система авторизации
- **Роли пользователей**: `user` (обычный пользователь), `admin` (администратор)
- **Региональная привязка**: каждый пользователь привязан к определенному региону
- **Сессионная авторизация**: использование cookie-based сессий

### Управление статистикой
Система отслеживает следующие показатели:
- **Семечка**: план, факт, отклонение
- **Тыква**: план, факт, отклонение  
- **Арахис**: план, факт, отклонение
- **Дополнительные метрики**: AKB1, AKB2, NewTT, Mix, NpOne, SetShel, DMP, TopFive, News

### Основные возможности
- ✅ Ввод ежедневной статистики
- ✅ Редактирование данных за текущий день
- ✅ Просмотр статистики по месяцам (до 30 дней назад)
- ✅ Автоматическое вычисление отклонений (факт - план)
- ✅ Агрегированная статистика по регионам
- ✅ Автоматическая очистка данных старше 30 дней

## 🚀 Как запустить проект

### Предварительные требования
- Go 1.23+
- Git
- Docker
- Docker Compose

### Настройка окружения
Создайте `.env` файл с переменными окружения:

```bash
# TRUNCATE INTERVAL
TRUNCATE_INTERVAL=5m

# SERVER
SVR_PORT=:47291

# DEBUG
DEBUG=true

# PASSWORD
PASSWORD=your_password

# COOKIE
SESSION_SECRET=session_key

# DB
DB_HOST=db
DB_USER=app_user
DB_PASSWORD=password
DB_NAME=app_db
DB_PORT=5432
```

### Запуск
```bash
# Клонируем репозиторий
git clone https://github.com/Wladim1r/Djin.git
cd Djin

# Устанавливаем зависимости
go mod tidy

# Запускаем приложение
make doc-up-n-c
# или
docker compose build --no-cache && docker compose up
```

## 🖥️ Интерфейс пользователя

### Для обычных пользователей:
- **Dashboard** (`/dashboard`) - главная страница
- **Ввод статистики** (`/inputStat`) - форма для ввода данных
- **Просмотр за месяц** (`/monthStat`) - статистика по месяцам

### Для администраторов:
- **Админ-панель** (`/admin/panel`) - управление пользователями
- **Все возможности обычных пользователей**

## 🔐 API Endpoints

### Авторизация
```http
POST /auth/login      # Вход в систему
POST /auth/logout     # Выход из системы
GET  /auth/me         # Информация о текущем пользователе
```

### Статистика
```http
POST   /djin/stat     # Создание/обновление статистики
PATCH  /djin/stat     # Редактирование статистики
GET    /djin/stat     # Получение статистики за сегодня
GET    /djin/month    # Статистика за определенную дату
GET    /djin/total    # Агрегированная статистика по региону
```

### Администрирование
```http
GET    /admin/users        # Список всех пользователей
POST   /admin/users        # Создание пользователя
PUT    /admin/users/:id    # Обновление пользователя
DELETE /admin/users/:id    # Удаление пользователя
GET    /admin/regions      # Список регионов
```

## 🔧 Особенности реализации

### Безопасность
- Middleware для проверки авторизации
- Роль-based доступ к функциям
- CORS настройки
- Безопасные заголовки HTTP
- Хеширование паролей с bcrypt

### Производительность
- Автоматическая очистка старых данных (30+ дней)
- In-memory агрегация статистики для быстрого доступа
- Индексы базы данных для оптимизации запросов

### Надежность
- Graceful shutdown сервера
- Контекстное управление горутинами
- Обработка ошибок на всех уровнях
- Логирование операций

## 📝 Модель данных

### StatDaily
```go
type StatDaily struct {
    ID       uint    `json:"id"`
    Date     string  `json:"date"`
    Name     string  `json:"name"`
    RegionID uint    `json:"region_id"`
    
    SeedPlan    float64 `json:"seed_plan"`
    SeedFact    float64 `json:"seed_fact"`
    SeedDif     float64 `json:"seed_dif"`
    
    PumpkinPlan float64 `json:"pumpkin_plan"`
    PumpkinFact float64 `json:"pumpkin_fact"`
    PumpkinDif  float64 `json:"pumpkin_dif"`
    
    PeanutPlan  float64 `json:"peanut_plan"`
    PeanutFact  float64 `json:"peanut_fact"`
    PeanutDif   float64 `json:"peanut_dif"`
    
    // Дополнительные метрики
	  AKB1    int `json:"akb1,omitempty"`
	  AKB2    int `json:"akb2,omitempty"`
	  NewTT   int `json:"newtt,omitempty"`
	  Mix     int `json:"mix,omitempty"`
	  NpOne   int `json:"npone,omitempty"`
	  SetShel int `json:"set_shelving,omitempty"`
	  DMP     int `json:"dmp,omitempty"`
	  TopFive int `json:"top_five,omitempty"`
	  News    int `json:"news,omitempty"`
}
```

## 🎯 Целевая аудитория

Система предназначена для:
- **Региональных менеджеров** - ввод и контроль ежедневной статистики
- **Администраторов** - управление пользователями и системой
- **Аналитиков** - просмотр агрегированных данных по регионам

## 🔄 Процессы автоматизации

- **Ежедневная очистка**: автоматическое удаление данных старше 30 дней в полночь
- **Автовычисления**: отклонения (факт - план) вычисляются автоматически
- **Агрегация**: статистика по регионам обновляется в реальном времени

---

*Проект активно развивается и поддерживается. Для вопросов и предложений используйте GitHub Issues.*
