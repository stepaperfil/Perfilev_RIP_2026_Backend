# Л/р 4 — авторизация JWT, Redis и Swagger

⬅ [Общее описание проекта и список веток](../../tree/main)

## Что реализовано в этом спринте

- Добавлена аутентификация по **JWT** (`golang-jwt/jwt`): метод входа проверяет логин и bcrypt-хеш пароля и выдаёт токен для заголовка `Authorization: Bearer <токен>`.
- Выход реализован через **блеклист в Redis** (`go-redis/v8`): токен хранится в Redis до окончания срока действия и больше не принимается.
- Авторизация через middleware `WithAuthCheck`: без токена доступны только чтение (список, лента), вход и регистрация; методы создателя (черновик, добавление, публикация, удаление, лайк, выход) возвращают `401`.
- Создатель этапа и автор лайка подставляются из токена (автозаполнение `creator_id`), функция-singleton из ЛР3 удалена.
- Описание методов в **Swagger** (`swaggo/swag`, `gin-swagger`), документация генерируется из комментариев: `http://localhost:8080/swagger/index.html`.
- Бэкенд работает как **REST API**: все методы отдают JSON и начинаются с `/api`.
- Изображения и видео этапов загружаются как файлы и хранятся в **MinIO**; в БД сохраняются только сгенерированные латинские имена файлов (`image_key`, `video_key`).
- Работа с БД только через **GORM**; удалённые этапы клиенту не передаются, удаление логическое.
- Статусы меняются только вперёд: `черновик` → `опубликован` → `удален`.
- Системные поля (id, статус, создатель, даты) с клиента не принимаются — они вычисляются на бэкенде.
- Пароль при регистрации хранится в виде bcrypt-хеша.

## Запуск

1. Поднять PostgreSQL, Adminer, Redis и MinIO:
   ```bash
   docker compose -f docker-compose.yml up -d
   docker compose -f docker-compose.minio.yml up -d
   ```
2. Создать базу ЛР4 (один раз): `docker exec tco-postgres createdb -U tco tco_lab4`
3. Миграции: `go run ./cmd/migrate`
4. Документация Swagger (после изменения комментариев): `swag init -g cmd/tcoApp/main.go -o docs`
5. Сервер: `go run ./cmd/tcoApp` (порт `8080`, можно задать переменной `PORT`)

Настройки читаются из `.env` (библиотека `godotenv`), при его отсутствии используются значения по умолчанию:

| Переменная | По умолчанию |
|---|---|
| `POSTGRES_HOST` / `POSTGRES_PORT` | `localhost` / `5432` |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` | `tco` / `tco` |
| `POSTGRES_DB` | `tco_lab4` |
| `MINIO_ENDPOINT` | `localhost:9000` |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `minioadmin` / `minioadmin` |
| `MINIO_BUCKET` | `tco-assets` |
| `REDIS_HOST` / `REDIS_PORT` | `localhost` / `6379` |
| `REDIS_PASSWORD` | `password` |
| `JWT_SECRET` | `tco-secret` |
| `JWT_EXPIRES_IN` | `1h` |

## Методы API

Ошибки возвращаются только кодом ответа без тела: `400` — некорректный запрос, `401` — нет действующего токена, `404` — не найдено, `409` — конфликт, `500` — ошибка сервера.

Доступ: **гость** — список, лента, регистрация, вход; **создатель** (с заголовком `Authorization: Bearer <токен>`) — все остальные методы.

### Этапы жизненного цикла (услуги)

| Метод | URL | Тело запроса | Ответ |
|---|---|---|---|
| GET | `/api/life_cycle_stages?max_price_per_unit=` | — | `200`, список опубликованных этапов; фильтр по цене за единицу (≤); у каждого этапа `likes_count` и признак `is_mine` (0/1 — создатель этапа совпадает с текущим пользователем) |
| GET | `/api/life_cycle_stages/feed?id=&next=true` | — | `200`, лента: без `id` — первый этап (по порядковому номеру), с `id` — этап `id`, а с `next=true` — следующий за ним; из БД читается одна строка; `404`, если этапа нет |
| GET | `/api/life_cycle_stages/draft` | — | `200`, черновик текущего пользователя (не более одного); `404`, если черновика нет |
| POST | `/api/life_cycle_stages` | `multipart/form-data`: `title`, `image` (файл jpg/png/webp/gif), `video` (файл mp4/webm) | `201`, созданный черновик; файлы загружаются в MinIO; `409`, если черновик уже есть |
| PUT | `/api/life_cycle_stages/:id/publish` | JSON: `description`, `price_per_unit`, `stage_position` | `200`, опубликованный этап; только свой черновик |
| DELETE | `/api/life_cycle_stages/:id` | — | `200`; логическое удаление (статус `удален`), только свои этапы |
| POST | `/api/life_cycle_stages/:id/like` | JSON: `like` — `1` ставит лайк, `0` отменяет | `200`, этап с обновлёнными `likes_count` и `is_liked` |

### Пользователь (инженер)

| Метод | URL | Тело запроса | Ответ |
|---|---|---|---|
| POST | `/api/engineers` | JSON: `full_name`, `login`, `password` | `201`, регистрация; `409`, если логин занят |
| POST | `/api/engineers/login` | JSON: `login`, `password` | `200`, `access_token`, `token_type`, `expires_in`; `401`, если логин или пароль неверны |
| POST | `/api/engineers/logout` | — (заголовок `Authorization`) | `200`, токен добавлен в блеклист Redis |

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
| `image_key` | varchar(255), not null | имя файла изображения в MinIO |
| `video_key` | varchar(255), not null | имя файла видео в MinIO |
| `price_per_unit` | bigint, null | цена за единицу, ₽ (поле фильтрации); заполняется при публикации |
| `stage_position` | bigint, null | порядковый номер этапа; заполняется при публикации |
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
- **JWT** (`golang-jwt/jwt`) — аутентификация и авторизация
- **Redis** (`go-redis/v8`) — блеклист токенов после выхода
- **Swagger** (`swaggo/swag`, `gin-swagger`) — документация API
- **Docker Compose** — локальный запуск PostgreSQL, Redis, MinIO и Adminer
