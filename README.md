# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из migrations/0001_create_tasks.up.sql монтируется в docker-entrypoint-initdb.d и применяется только при инициализации пустого data volume. (добавлен файл 0002_add_task_recurrence.up.sql для примеров с периодичностью) 

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON: (обновлен с примерами запросов для периодичных задач)

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Добавлено: периодичность задач

Поддерживаемые типы:

- `none` — без расписания;
- `daily` — каждый `interval_days` день (`interval_days >= 1`)
- `monthly` — в `day_of_month` число месяца (`1..30`)
- `parity` — только четные/нечетные дни месяца (`is_even=true|false`)
- `dates` — только в даты из `specific_dates` (формат `YYYY-MM-DD`)

### Обязательные для каждого типа

- `none`: -
- `daily`: `interval_days`
- `monthly`: `day_of_month`
- `parity`: `is_even`
- `dates`: not nil `specific_dates`

## Вычисление следующего дня активности задачи (`next_run_at`)

Используем UTC и начало суток (`00:00:00Z`) для расчета очередного запуска:

- `none`: `next_run_at = nil`
- `daily`: `next_run_at = today_utc + interval_days`
- `monthly`: ближайшая дата с указанным числом месяца (текущий или следующий месяц)
- `parity`: ближайший день с нужной четностью (включая сегодня)
- `dates`: ближайшая дата из `specific_dates`, которая не раньше текущего дня UTC

## Граничные случаи

- Пустой/невалидный `interval_type`
  - пустой трактуется как `nil`
  - невалидный → `400`
- Для `daily/monthly/parity/dates` отсутствие обязательного поля - `400`
- Для `dates`:
  - пустые строки в массиве отбрасываются
  - даты нормализуются к `YYYY-MM-DD`
  - дубликаты удаляются
  - если все даты в прошлом — `400`
- Для `monthly` если задача стоит на 31 число, в следующем месяце она будет поставлена на 30-е (в случае февраля на 28/29), в случае июль->август где оба месяца есть 31 число, задача останется 31-го

## Ограничения

- Расчет дат в UTC

## Особенности

- Задачи продлеваются два раза в день - в 6 и 18 часов с помощью планировщика (запускается как горутина в main)
- Даты для `specific_dates` принимаем в формате `YYYY-MM-DD`
- Для четности считаем, что `is_even=true` означает запуск по четным числам месяца, `false` — по нечетным
- Для `none` любые дополнительные поля периодичности игнорируем

