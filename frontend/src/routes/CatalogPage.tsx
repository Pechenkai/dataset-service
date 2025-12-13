import React from 'react';
import { Link } from 'react-router-dom';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { useCatalogViewModel } from '../viewmodels/catalogViewModel';
import { Category, Dataset } from '../api/types';

const DatasetCard: React.FC<{ dataset: Dataset; category?: Category }> = ({ dataset, category }) => {
  const tagList = dataset.metadata?.tags || [];
  return (
    <Card
      title={dataset.name}
      subtitle={category ? category.name : 'Без категории'}
      toolbar={<Badge tone={dataset.is_public ? 'success' : 'warning'}>{dataset.is_public ? 'Публичный' : 'Приватный'}</Badge>}
    >
      <p>{dataset.description || 'Описание пока пустое — самое время его добавить.'}</p>
      <div className="meta-bar">
        <span>Версия: {dataset.latest_version?.number || '—'}</span>
        {dataset.rating_summary && (
          <span>
            Рейтинг: {dataset.rating_summary.average.toFixed(1)} ({dataset.rating_summary.count})
          </span>
        )}
      </div>
      {tagList.length > 0 && (
        <div className="tag-list">
          {tagList.map((tag) => (
            <span className="tag" key={tag}>
              #{tag}
            </span>
          ))}
        </div>
      )}
      <Link className="ui-button ui-button--outline" to={`/datasets/${dataset.id}`}>
        Открыть карточку
      </Link>
    </Card>
  );
};

const CatalogPage: React.FC = () => {
  const { categories, datasets, filters, meta, isFetching, updateFilters } = useCatalogViewModel();
  const totalPages = meta ? Math.max(1, Math.ceil(meta.total / meta.limit)) : 1;

  return (
    <>
      <section className="page-hero">
        <h2>Каталог датасетов</h2>
        <p>Фильтры синхронизируются с URL, чтобы можно было делиться ссылками на конкретный срез.</p>
      </section>

      <Card title="Поиск и фильтры" subtitle="Все изменения сразу попадают в адресную строку">
        <div className="filters-grid">
          <Input
            label="Поиск"
            placeholder="Название или описание"
            value={filters.search || ''}
            onChange={(e) => updateFilters({ search: e.target.value, page: 1 })}
          />
          <label className="ui-input">
            <span className="ui-input__label">Категория</span>
            <select
              className="ui-input__field"
              value={filters.categoryId || ''}
              onChange={(e) => updateFilters({ categoryId: e.target.value ? Number(e.target.value) : undefined, page: 1 })}
            >
              <option value="">Все</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.id}>
                  {cat.name}
                </option>
              ))}
            </select>
          </label>
          <label className="ui-input">
            <span className="ui-input__label">Доступность</span>
            <select
              className="ui-input__field"
              value={filters.visibility || 'all'}
              onChange={(e) => updateFilters({ visibility: e.target.value as any, page: 1 })}
            >
              <option value="all">Публичные и приватные</option>
              <option value="public">Только публичные</option>
              <option value="private">Только приватные</option>
            </select>
          </label>
          <Input
            label="Теги"
            placeholder="computer-vision,nlp"
            value={filters.tags || ''}
            onChange={(e) => updateFilters({ tags: e.target.value, page: 1 })}
            hint="Разделение запятыми"
          />
        </div>
      </Card>

      <div className="meta-bar">
        <span>{isFetching ? 'Обновляем список…' : `Найдено: ${meta?.total ?? datasets.length}`}</span>
        <span>Страница {filters.page ?? 1} из {totalPages}</span>
      </div>

      <div className="catalog-grid">
        {datasets.map((dataset) => (
          <DatasetCard
            key={dataset.id}
            dataset={dataset}
            category={categories.find((cat) => cat.id === dataset.category_id)}
          />
        ))}
        {datasets.length === 0 && <Card title="Нет данных">Подходящих датасетов не найдено.</Card>}
      </div>

      {totalPages > 1 && (
        <div className="meta-bar" style={{ justifyContent: 'space-between' }}>
          <Button
            variant="ghost"
            onClick={() => updateFilters({ page: Math.max(1, (filters.page || 1) - 1) })}
            disabled={(filters.page || 1) <= 1}
          >
            Назад
          </Button>
          <Button
            variant="ghost"
            onClick={() => updateFilters({ page: Math.min(totalPages, (filters.page || 1) + 1) })}
            disabled={(filters.page || 1) >= totalPages}
          >
            Вперед
          </Button>
        </div>
      )}
    </>
  );
};

export default CatalogPage;
