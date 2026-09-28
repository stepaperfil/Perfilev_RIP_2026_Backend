# Л/р 3 — веб-сервис REST API

⬅ [Общее описание проекта и список веток](../../tree/main)

## Что реализовано в этом спринте

- Бэкенд переведён с HTML-шаблонов на **REST API**: все методы отдают JSON и начинаются с `/api`, тестирование через Postman/Insomnia.
- Изображения и видео этапов загружаются как файлы и хранятся в **MinIO**; в БД сохраняются только сгенерированные латинские имена файлов (`image_key`, `video_key`).
- Работа с БД только через **GORM**; удалённые этапы клиенту не передаются, удаление логическое.
- Статусы меняются только вперёд: `черновик` → `опубликован` → `удален`.
- Пользователь-создатель зафиксирован константой через функцию-singleton `CurrentEngineerID()` (до авторизации в ЛР4).
- Системные поля (id, статус, создатель, даты) с клиента не принимаются — они вычисляются на бэкенде.
- Пароль при регистрации хранится в виде bcrypt-хеша.

## Запуск

1. Поднять PostgreSQL, Adminer и MinIO:
   ```bash
   docker compose -f docker-compose.yml up -d
   docker compose -f docker-compose.minio.yml up -d
   ```
2. Создать базу ЛР3 (один раз): `docker exec tco-postgres createdb -U tco tco_lab3`
3. Миграции: `go run ./cmd/migrate`
4. Сервер: `go run ./cmd/tcoApp` (порт `8080`, можно задать переменной `PORT`)

Настройки читаются из `.env` (библиотека `godotenv`), при его отсутствии используются значения по умолчанию:

| Переменная | По умолчанию |
|---|---|
| `POSTGRES_HOST` / `POSTGRES_PORT` | `localhost` / `5432` |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` | `tco` / `tco` |
| `POSTGRES_DB` | `tco_lab3` |
| `MINIO_ENDPOINT` | `localhost:9000` |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `minioadmin` / `minioadmin` |
| `MINIO_BUCKET` | `tco-assets` |

## Методы API

Ошибки возвращаются только кодом ответа без тела: `400` — некорректный запрос, `404` — не найдено, `409` — конфликт, `500` — ошибка сервера.

### Этапы жизненного цикла (услуги)

| Метод | URL | Тело запроса | Ответ |
|---|---|---|---|
| GET | `/api/life_cycle_stages?max_price_per_unit=` | — | `200`, список опубликованных этапов; фильтр по цене за единицу (≤); у каждого этапа `likes_count` и признак `is_mine` (0/1 — создатель этапа совпадает с текущим пользователем) |
| GET | `/api/life_cycle_stages/feed` | — | `200`, первый этап ленты (по порядковому номеру) |
| GET | `/api/life_cycle_stages/feed/:id?next=true` | — | `200`, этап `:id`, а с `next=true` — следующий за ним; из БД читается одна строка; `404`, если этапа нет |
| GET | `/api/life_cycle_stages/draft` | — | `200`, черновик текущего пользователя (не более одного); `404`, если черновика нет |
| POST | `/api/life_cycle_stages` | `multipart/form-data`: `title`, `image` (файл jpg/png/webp/gif), `video` (файл mp4/webm) | `201`, созданный черновик; файлы загружаются в MinIO; `409`, если черновик уже есть |
| PUT | `/api/life_cycle_stages/:id/publish` | JSON: `description`, `price_per_unit`, `stage_position` | `200`, опубликованный этап; только свой черновик |
| DELETE | `/api/life_cycle_stages/:id` | — | `200`; логическое удаление (статус `удален`), только свои этапы |
| POST | `/api/life_cycle_stages/:id/like` | JSON: `like` — `1` ставит лайк, `0` отменяет | `200`, этап с обновлёнными `likes_count` и `is_liked` |

### Пользователь (инженер)

| Метод | URL | Тело запроса | Ответ |
|---|---|---|---|
| POST | `/api/engineers` | JSON: `full_name`, `login`, `password` | `201`, регистрация; `409`, если логин занят |
| POST | `/api/engineers/login` | JSON: `login`, `password` | `200`, заглушка до ЛР4 |
| POST | `/api/engineers/logout` | — | `200`, заглушка до ЛР4 |

## Таблицы БД

### `engineers` — пользователи

| Поле | Тип | Описание |
|---|---|---|
| `id` | bigint, PK | идентификатор |
| `full_name` | varchar(150) | ФИО |
| `login` | varchar(100), unique | логин |
| `password_hash` | varchar(255) | bcrypt-хеш пароля |

### `life_cycle_stages` — этапы жизненного цикла актива (услуги)

| Поле | Тип | Описание |
|---|---|---|
| `id` | bigint, PK | идентификатор |
| `title` | varchar(150) | название |
| `description` | text | краткое описание |
| `status` | varchar(20) | `черновик` / `опубликован` / `удален` |
| `image_key` | varchar(255) | имя файла изображения в MinIO |
| `video_key` | varchar(255) | имя файла видео в MinIO |
| `price_per_unit` | bigint | цена за единицу, ₽ (поле фильтрации) |
| `stage_position` | bigint | порядковый номер этапа |
| `created_at` | timestamptz | дата создания |
| `formed_at` | timestamptz | дата формирования (публикации) |
| `creator_id` | bigint, FK → `engineers.id` | создатель |

### `life_cycle_stage_likes` — лайки (м-м инженеры ↔ этапы)

| Поле | Тип | Описание |
|---|---|---|
| `id` | bigint, PK | идентификатор |
| `engineer_id` | bigint, FK → `engineers.id` | кто поставил лайк |
| `life_cycle_stage_id` | bigint, FK → `life_cycle_stages.id` | этап |

Пара (`engineer_id`, `life_cycle_stage_id`) уникальна.

## Технологии

- **Go** + **Gin** — REST API
- **PostgreSQL** + **GORM** — хранение данных и ORM
- **MinIO** (`minio-go/v7`) — хранение изображений и видео
- **godotenv** — настройки из `.env`
- **bcrypt** (`golang.org/x/crypto`) — хеширование паролей
- **Docker Compose** — локальный запуск PostgreSQL, MinIO и Adminer
