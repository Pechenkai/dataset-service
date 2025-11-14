# Имитация E2E-сценария и снятие трафика

Этот документ описывает, как воспроизвести пользовательский сценарий из `internal/tests/e2e/user_journey_e2e_test.go` и зафиксировать сетевой трафик для предъявления преподавателю.

## Инструменты

- **Postman** — основное средство отправки HTTP-запросов. В репозитории есть готовая коллекция `docs/postman/public_dataset_journey.postman_collection.json`.
- **curl** — альтернативный CLI-клиент (используется в автоматизированном скрипте).
- **Wireshark / tcpdump** — захват и анализ HTTP-трафика (фильтр `tcp.port == 8080`).

## Подготовка окружения

1. Поднимите полное окружение (Postgres, Mongo, MinIO, API):
   ```bash
   docker compose up -d --build db mongo minio api
   ```

   Контейнер `api` уже содержит приложение из `cmd/api` и слушает `http://localhost:8080`.

2. Выполните `go run ./cmd/e2e_seed` и зафиксируйте напечатанные `USER_ID`, `DATASET_ID` и `CATEGORY_ID`. Эти значения понадобятся в Postman.

3. Очистите локальные сессии/куки в Postman, чтобы запросы выполнялись «с нуля».

## Ручная последовательность запросов (Postman)

1. Импортируйте коллекцию `docs/postman/public_dataset_journey.postman_collection.json`.
2. Создайте окружение с переменными:
   - `baseUrl` — `http://localhost:8080`
   - `userId` / `datasetId` — значения из вывода `cmd/e2e_seed`
3. Запустите коллекцию в Runner либо пошагово. Запросы внутри коллекции отвечают шагам `GET /categories`, `GET /datasets?public=true`, `GET /datasets/{id}/versions`, `POST /subscriptions`, `POST /reviews`.
4. Включите `tcpdump`/Wireshark и зафиксируйте пакеты. Для Wireshark удобно применить фильтр `tcp.port == 8080`.

## Автоматизация через скрипт

Чтобы не выполнять шаги вручную, добавлена цель `make e2e-capture`, которая:

1. Запускает `go run ./cmd/e2e_seed` и создаёт тестовые сущности (user/category/dataset) в той же базе и MinIO, что использует API.
2. Стартует `tcpdump` (интерфейс `lo`, порт `8080`) и пишет расшифровку HTTP-пакетов в `logs/e2e_capture_example.txt`.
3. Выполняет ту же последовательность запросов через `curl`.
4. Останавливает захват и обновляет лог.

Если `tcpdump` требует root, выполните `sudo make e2e-capture`. Кастомизация через переменные окружения:
- `API_BASE` — адрес API (по умолчанию `http://localhost:8080`);
- `INTERFACE` / `API_PORT` — интерфейс и порт для `tcpdump`;
- `USER_ID`, `DATASET_ID` — можно переопределить при необходимости.

## Включение Wireshark

Для сохранения PCAP-файла, который можно открыть в Wireshark, используйте цель:
```bash
make e2e-wireshark
# или, если tcpdump требует пароль:
sudo make e2e-wireshark
```

Команда выполнит те же шаги, но дополнительно создаст `logs/e2e_capture.pcap`. Откройте этот файл в Wireshark, примените фильтр `tcp.port == 8080` и покажите преподавателю последовательность `GET /categories`, `POST /reviews`, …, `POST /subscriptions`. Текстовый лог `logs/e2e_capture_example.txt` также будет обновлён ASCII-представлением дампа.
