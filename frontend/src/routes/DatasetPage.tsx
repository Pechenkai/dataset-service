import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { useDatasetViewModel } from '../viewmodels/datasetViewModel';
import { useServices } from '../context/ServiceContext';

const DatasetPage: React.FC = () => {
  const { datasetId } = useParams<{ datasetId: string }>();
  const id = Number(datasetId);
  const { datasetService } = useServices();
  const categoriesQuery = useQuery(['categories'], () => datasetService.listCategories());
  const { datasetQuery, versionsQuery, reviewsQuery, createReview, subscribe } = useDatasetViewModel(id);
  const [rating, setRating] = useState(5);
  const [text, setText] = useState('');

  if (datasetQuery.isLoading) {
    return <p>Загружаем датасет…</p>;
  }

  if (!datasetQuery.data) {
    return <p>Датасет не найден.</p>;
  }

  const dataset = datasetQuery.data;
  const categoryName = categoriesQuery.data?.items.find((c) => c.id === dataset.category_id)?.name;

  return (
    <>
      <section className="page-hero">
        <h2>{dataset.name}</h2>
        <p>{dataset.description || 'Описание появится позже.'}</p>
        <div className="meta-bar">
          <Badge tone={dataset.is_public ? 'success' : 'warning'}>
            {dataset.is_public ? 'Публичный' : 'Приватный'}
          </Badge>
          {categoryName && <span>Категория: {categoryName}</span>}
          {dataset.latest_version && <span>Текущая версия: {dataset.latest_version.number}</span>}
        </div>
      </section>

      <Card title="Информация">
        <div className="form-grid">
          <div className="stat-tile">
            <span className="stat-tile__label">Владелец</span>
            <span className="stat-tile__value">#{dataset.owner_id}</span>
          </div>
          <div className="stat-tile">
            <span className="stat-tile__label">Создан</span>
            <span className="stat-tile__value">{dataset.created_at?.slice(0, 10) || '—'}</span>
          </div>
          <div className="stat-tile">
            <span className="stat-tile__label">Формат</span>
            <span className="stat-tile__value">{dataset.metadata?.format || 'не указан'}</span>
          </div>
          <div className="stat-tile">
            <span className="stat-tile__label">Размер</span>
            <span className="stat-tile__value">{dataset.metadata?.size ? `${Math.round(dataset.metadata.size / 1024 / 1024)} МБ` : '—'}</span>
          </div>
        </div>
        <Button onClick={() => subscribe.mutate()}>Подписаться на обновления</Button>
      </Card>

      <Card title="Версии" subtitle="Последние сборки датасета">
        {versionsQuery.data?.items.map((version) => (
          <div key={version.id} className="stat-tile">
            <div className="meta-bar">
              <strong>{version.number}</strong>
              <span>{version.upload_date?.slice(0, 10) || '—'}</span>
            </div>
            <p>{version.change_log || 'Без описания изменений.'}</p>
            {version.metadata?.tags && (
              <div className="tag-list">
                {version.metadata.tags.map((tag) => (
                  <span className="tag" key={tag}>
                    #{tag}
                  </span>
                ))}
              </div>
            )}
            {version.file_url && (
              <a className="ui-button ui-button--outline" href={version.file_url} target="_blank" rel="noreferrer">
                Скачать версию
              </a>
            )}
          </div>
        ))}
        {versionsQuery.data?.items.length === 0 && <p>Версии пока не опубликованы.</p>}
      </Card>

      <Card title="Отзывы" subtitle="Оценка помогает владельцу понять ценность датасета">
        <div className="form-grid">
          <label className="ui-input">
            <span className="ui-input__label">Оценка</span>
            <input
              className="ui-input__field"
              type="number"
              min={1}
              max={5}
              value={rating}
              onChange={(e) => setRating(Number(e.target.value))}
            />
          </label>
          <label className="ui-input">
            <span className="ui-input__label">Комментарий</span>
            <textarea
              className="ui-input__field"
              rows={3}
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Что понравилось или не хватило"
            />
          </label>
          <Button onClick={() => createReview.mutate({ rating, text })} disabled={createReview.isLoading}>
            Опубликовать отзыв
          </Button>
        </div>

        <div className="catalog-grid">
          {reviewsQuery.data?.items.map((review) => (
            <div key={review.id} className="stat-tile">
              <div className="meta-bar">
                <Badge tone="info">{review.rating} / 5</Badge>
                <span>Пользователь #{review.user_id}</span>
                <span>{review.created_at?.slice(0, 10) || '—'}</span>
              </div>
              <p>{review.text || 'Без комментариев'}</p>
            </div>
          ))}
          {reviewsQuery.data?.items.length === 0 && <p>Пока нет отзывов.</p>}
        </div>
      </Card>
    </>
  );
};

export default DatasetPage;
