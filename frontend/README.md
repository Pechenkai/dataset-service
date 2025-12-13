# Dataset Platform SPA

Одностраничное приложение на React + TypeScript поверх Vite. Слои разделены на:

- `src/api` — HTTP-клиент, типы OpenAPI, QueryClient.
- `src/services` — бизнес-логика и работа с API (datasets, auth, notifications, subscriptions).
- `src/viewmodels` — view-model hooks (MVVM), связывают сервисы, состояние URL и представление.
- `src/routes` — страницы/экраны, маршрутизация React Router.
- `src/components` — переиспользуемые UI-компоненты и лэйауты.

## Скрипты

```bash
npm install          # установка зависимостей
npm run dev          # запуск Vite на http://localhost:4173/app/
npm run build        # сборка SPA в ../static/app
npm run test         # unit/integration тесты (vitest)
npm run test:coverage# отчёт покрытия (порог 60%+)
```

## Особенности

- Базовый путь роутера `/app`, сборка складывается в `static/app` (игнорируется Git).
- Авторизация вынесена в `AuthService`, токен хранится в localStorage, добавляется в ApiClient через токен-провайдер.
- Каталог синхронизирует фильтры с URL (параметры `q`, `category`, `visibility`, `tags`, `page`).
- Формы публикации/версий используют `FormData` и мапят поля на OpenAPI v2.
- Тестовый стек: Vitest + Testing Library, покрытие с порогами в `vitest.config.ts`.
