import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { useServices } from '../context/ServiceContext';

const UploadPage: React.FC = () => {
  const { datasetService } = useServices();
  const categoriesQuery = useQuery(['categories'], () => datasetService.listCategories());
  const [message, setMessage] = useState('');
  const [datasetForm, setDatasetForm] = useState({
    name: '',
    description: '',
    categoryId: '',
    isPublic: true,
    format: 'csv',
    tags: 'ml,experiment',
    file: null as File | null
  });

  const disableSubmit = !datasetForm.name || !datasetForm.categoryId || !datasetForm.file;

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!datasetForm.file) return;
    await datasetService.createDataset({
      name: datasetForm.name,
      description: datasetForm.description,
      categoryId: Number(datasetForm.categoryId),
      isPublic: datasetForm.isPublic,
      file: datasetForm.file,
      metadata: {
        format: datasetForm.format,
        tags: datasetForm.tags.split(',').map((tag) => tag.trim()).filter(Boolean)
      }
    });
    setMessage('Датасет передан в API. После ответа сервера он появится в каталоге.');
  };

  const fileLabel = useMemo(() => datasetForm.file?.name || 'Файл не выбран', [datasetForm.file]);

  return (
    <>
      <section className="page-hero">
        <h2>Публикация датасета</h2>
        <p>Форма собирает payload для эндпоинта POST /api/v2/datasets. Файлы уходят через FormData.</p>
      </section>

      <Card title="Новый датасет" subtitle="Бизнес-логика вынесена в DatasetService">
        <form className="form-grid" onSubmit={handleSubmit}>
          <Input
            label="Название"
            placeholder="Traffic logs dataset"
            value={datasetForm.name}
            onChange={(e) => setDatasetForm({ ...datasetForm, name: e.target.value })}
            required
          />
          <Input
            label="Описание"
            placeholder="Пара слов о содержании набора"
            value={datasetForm.description}
            onChange={(e) => setDatasetForm({ ...datasetForm, description: e.target.value })}
          />
          <label className="ui-input">
            <span className="ui-input__label">Категория</span>
            <select
              className="ui-input__field"
              value={datasetForm.categoryId}
              onChange={(e) => setDatasetForm({ ...datasetForm, categoryId: e.target.value })}
              required
            >
              <option value="" disabled>
                Выберите категорию
              </option>
              {categoriesQuery.data?.items.map((cat) => (
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
              value={datasetForm.isPublic ? 'public' : 'private'}
              onChange={(e) => setDatasetForm({ ...datasetForm, isPublic: e.target.value === 'public' })}
            >
              <option value="public">Публичный</option>
              <option value="private">Приватный</option>
            </select>
          </label>
          <Input
            label="Формат"
            value={datasetForm.format}
            onChange={(e) => setDatasetForm({ ...datasetForm, format: e.target.value })}
          />
          <Input
            label="Теги"
            value={datasetForm.tags}
            onChange={(e) => setDatasetForm({ ...datasetForm, tags: e.target.value })}
            hint="Через запятую"
          />
          <label className="ui-input">
            <span className="ui-input__label">Файл</span>
            <input
              className="ui-input__field"
              type="file"
              onChange={(e) => setDatasetForm({ ...datasetForm, file: e.target.files?.[0] || null })}
              required
            />
            <span className="ui-input__hint">{fileLabel}</span>
          </label>
          <Button type="submit" disabled={disableSubmit}>
            Отправить в API
          </Button>
          {message && <p>{message}</p>}
        </form>
      </Card>
    </>
  );
};

export default UploadPage;
