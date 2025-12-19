import React, { useMemo } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Card } from '../components/ui/Card';
import { Table } from '../components/ui/Table';
import { useAuth } from '../context/AuthContext';
import { useServices } from '../context/ServiceContext';
import { DatasetFilters } from '../services/datasetService';
import { Dataset } from '../api/types';
import { formatSize } from './CatalogPage';
import { Button } from '../components/ui/Button';

const MyDatasetsPage: React.FC = () => {
  const { session } = useAuth();
  const { datasetService } = useServices();

  const ownerId = session?.user.id;

  const filters: DatasetFilters = useMemo(() => ({ ownerId, limit: 50 }), [ownerId]);
  const categoriesQuery = useQuery({
    queryKey: ['categories'],
    queryFn: () => datasetService.listCategories()
  });

  const myQuery = useQuery({
    queryKey: ['my-datasets', ownerId],
    queryFn: () => datasetService.listDatasets(filters),
    enabled: Boolean(ownerId)
  });

  const toggleVisibility = useMutation({
    mutationFn: (d: Dataset) => datasetService.updateDataset(d.id as number, { is_public: !d.is_public }),
    onSuccess: () => myQuery.refetch()
  });

  if (!ownerId) {
    return <Card title="Требуется вход">Авторизуйтесь, чтобы увидеть свои датасеты.</Card>;
  }

  return (
    <>
      <section className="page-hero">
        <h2>Мои датасеты</h2>
        <p>Личный список датасетов.</p>
        <Link to="/upload">
          <span className="ui-button ui-button--primary ui-button--md">Добавить датасет</span>
        </Link>
      </section>

      <Card title="Список" subtitle="Только ваши датасеты">
        {myQuery.data?.items.length ? (
          <Table<Dataset>
            rowKey={(d) => d.id}
            rows={myQuery.data?.items ?? []}
            gridTemplate="2fr 1.2fr 0.8fr 0.8fr 0.8fr 140px"
            columns={[
              {
                key: 'name',
                title: 'Название',
                render: (d) => <Link to={`/datasets/${d.id}`}>{d.name}</Link>
              },
              {
                key: 'category',
                title: 'Категория',
                render: (d) => categoriesQuery.data?.items.find((c) => c.id === d.category_id)?.name || '—'
              },
              {
                key: 'updated',
                title: 'Обновлено',
                render: (d) => d.updated_at?.slice(0, 10) || d.created_at?.slice(0, 10) || '—'
              },
              {
                key: 'rating',
                title: 'Рейтинг',
                render: (d) => d.rating_summary ? `${d.rating_summary.average.toFixed(1)} (${d.rating_summary.count})` : '—'
              },
              {
                key: 'size',
                title: 'Размер',
                render: (d) =>
                  formatSize(
                    d.metadata?.size ??
                    d.latest_version?.metadata?.size ??
                    (d.latest_version as any)?.size
                  ) ?? '—'
              },
              {
                key: 'visibility',
                title: 'Публичность',
                render: (d) => (
                  <Button
                    size="sm"
                    variant={d.is_public ? 'outline' : 'ghost'}
                    onClick={() => toggleVisibility.mutate(d)}
                    disabled={toggleVisibility.isLoading}
                  >
                    {d.is_public ? 'Публичный' : 'Приватный'}
                  </Button>
                )
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
        ) : (
          <p>Пока пусто.</p>
        )}
      </Card>
    </>
  );
};

export default MyDatasetsPage;
