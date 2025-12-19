## Асинхронная декомпозиция (Gateway/Core/Data + RabbitMQ)

- Транспорт: RabbitMQ (`gateway.core.cmd`, `core.gateway.evt`, `core.data.cmd`, `data.core.evt`), durable очереди, prefetch задаётся через `BROKER_PREFETCH`.
- Сервисы: `gateway` (HTTP API на `/async/*`, прокси в Rabbit), `core` (оркестрация бизнес-команд), `data` (доступ к БД). Каждый сервис можно масштабировать `docker compose up --scale gateway=3 --scale core=3 --scale data=3`.
- Метрики и здоровье: `/metrics` и `/healthz` на порте сервиса (по умолчанию `8080`). Prometheus таргеты добавлены в `deploy/monitoring/prometheus.yml`, Nginx пробрасывает `/async/` на gateway.
- Конфиг: блок `broker` в `config.yaml` и переменные окружения `BROKER_URL`, `BROKER_GATEWAY_TO_CORE_QUEUE` и т.д. `MessageProcessTimeout` управляет RPC-таймаутами.
- Демонстрационный сценарий: запрос категорий через асинхронную цепочку `GET http://localhost:8088/async/categories` → Gateway → Core → Data → Core → Gateway. Логи и метрики попадают в Loki/Prometheus.
- Дополнительные операции (асинхронные RPC): `GET /async/datasets?only_public=true&owner_id=...`, `GET /async/datasets/{id}`, `GET /async/datasets/{id}/versions`, `GET /async/versions/{id}`, `GET /async/notifications/{userID}`, `GET /async/access/owner/{ownerID}/pending`, `GET /async/access?dataset_id=..&user_id=..`.
