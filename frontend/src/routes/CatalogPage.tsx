import React from 'react';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Pagination } from '../components/ui/Pagination';
import { Table } from '../components/ui/Table';
import { Link } from 'react-router-dom';
import { useCatalogViewModel } from '../viewmodels/catalogViewModel';
import { Dataset } from '../api/types';
import { createApiClient } from '../api/client';

export const formatSize = (raw?: number | string | null) => {
  if (raw === undefined || raw === null) return null;
  const size = typeof raw === 'string' ? Number(raw) : raw;
  if (!Number.isFinite(size) || size <= 0) return null;
  if (size > 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  if (size > 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${size} B`;
};

const CatalogPage: React.FC = () => {
  const { categories, datasets, filters, meta, isFetching, updateFilters } = useCatalogViewModel();
  const totalPages = meta ? Math.max(1, Math.ceil(meta.total / meta.limit)) : 1;
  const currentPage = filters.page ?? 1;

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
        <span>
          Страница {currentPage} из {totalPages}
        </span>
      </div>

      <Card title="Список датасетов">
        {datasets.length === 0 ? (
          <div>Подходящих датасетов не найдено.</div>
        ) : (
          <Table<Dataset>
            rowKey={(d) => d.id}
            rows={datasets}
            gridTemplate="2fr 1.2fr 0.8fr 0.8fr 140px"
            columns={[
              {
                key: 'name',
                title: 'Название',
                render: (d) => <Link to={`/datasets/${d.id}`}>{d.name}</Link>
              },
              {
                key: 'category',
                title: 'Категория',
                render: (d) => categories.find((c) => c.id === d.category_id)?.name || '—'
              },
              {
                key: 'rating',
                title: 'Рейтинг',
                render: (d) => d.rating_summary ? `${d.rating_summary.average.toFixed(1)} (${d.rating_summary.count})` : '—'
              },
              {
                key: 'size',
                title: 'Размер',
                render: (d) => {
                  const size = formatSize(
                    d.metadata?.size ??
                    d.latest_version?.metadata?.size ??
                    (d.latest_version as any)?.size
                  );
                  return size ?? '—';
                }
              },
              {
                key: 'updated',
                title: 'Обновлено',
                render: (d) => d.updated_at?.slice(0, 10) || d.created_at?.slice(0, 10) || '—'
              },
              {
                key: 'download',
                title: '',
                render: (d) =>
                  d.is_public && d.latest_version?.id ? (
                    <a
                      href={`/api/v2/datasets/${d.id}/versions/${d.latest_version.id}/content`}
                      aria-label="Download"
                      className="app-link"
                      target="_blank"
                      rel="noreferrer"
                    >
                      ⬇
                    </a>
                  ) : (
                    '—'
                  )
              }
            ]}
          />
        )}
      </Card>

      <div style={{ marginTop: 16 }}>
        <Pagination page={currentPage} totalPages={totalPages} onPageChange={(p) => updateFilters({ page: p })} />
      </div>
    </>
  );
};

export default CatalogPage;
