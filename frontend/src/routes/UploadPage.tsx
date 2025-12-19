import React, { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { TextArea } from '../components/ui/TextArea';
import { Button } from '../components/ui/Button';
import { FileDropInput } from '../components/ui/FileDropInput';
import { useServices } from '../context/ServiceContext';
import { ApiError } from '../api/client';

const UploadPage: React.FC = () => {
  const { datasetService } = useServices();

  const categoriesQuery = useQuery({
    queryKey: ['categories'],
    queryFn: () => datasetService.listCategories()
  });

  const [form, setForm] = useState({
    name: '',
    tags: '',
    description: '',
    categoryId: '',
    file: null as File | null,
    format: 'csv',
    isPublic: true
  });

  const [status, setStatus] = useState<{ kind: 'idle' | 'ok' | 'err'; text?: string }>({ kind: 'idle' });

  useEffect(() => {
    const first = categoriesQuery.data?.items?.[0];
    if (first && !form.categoryId) {
      setForm((s) => ({ ...s, categoryId: String(first.id) }));
    }
  }, [categoriesQuery.data, form.categoryId]);

  const categoryId = useMemo(() => {
    const found = categoriesQuery.data?.items.find((c) => String(c.id) === form.categoryId);
    return found?.id ?? null;
  }, [categoriesQuery.data, form.categoryId]);

  const canSubmit = Boolean(form.name.trim() && form.file && categoryId);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setStatus({ kind: 'idle' });

    if (!categoryId) {
      setStatus({ kind: 'err', text: 'Нет категорий. Добавь хотя бы одну категорию в системе.' });
      return;
    }
    if (!form.file) return;

    try {
      await datasetService.createDataset({
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        categoryId: categoryId!,
        isPublic: form.isPublic,
        file: form.file,
        metadata: {
          format: form.format,
          tags: form.tags
              .split(',')
              .map((t) => t.trim())
              .filter(Boolean)
        }
      });
      setStatus({ kind: 'ok', text: 'Датасет создан.' });
      setForm({ name: '', tags: '', description: '', categoryId: '', file: null, format: 'csv', isPublic: true });
    } catch (err) {
      let message = 'Не удалось создать датасет.';
      if (err instanceof ApiError) {
        const body = err.body as any;
        message = body?.message || message;
      } else if (err instanceof Error) {
        message = err.message || message;
      }
      setStatus({ kind: 'err', text: message });
    }
  };

  return (
      <div className="create-ds">
        {/*<div className="create-ds__hint">Create dataset</div>*/}

        <Card title="">
          <form
              className="create-ds__form"
              onSubmit={submit}
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => e.preventDefault()}
          >
            <div className="create-ds__fields">
              <Input
                  label="Название"
                  placeholder="Input field"
                  value={form.name}
                  onChange={(e) => setForm((s) => ({ ...s, name: e.target.value }))}
                  required
              />

              <label className="ui-input">
                <span className="ui-input__label">Категория</span>
                <select
                    className="ui-input__field"
                    value={form.categoryId}
                    onChange={(e) => setForm((s) => ({ ...s, categoryId: e.target.value }))}
                    disabled={categoriesQuery.isLoading || !categoriesQuery.data?.items.length}
                    required
                >
                  {categoriesQuery.data?.items.map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </label>

              <Input
                  label="Теги"
                  placeholder="Input field"
                  value={form.tags}
                  onChange={(e) => setForm((s) => ({ ...s, tags: e.target.value }))}
                  hint="Через запятую"
              />

              <label className="ui-input">
                <span className="ui-input__label">Формат</span>
                <select
                    className="ui-input__field"
                    value={form.format}
                    onChange={(e) => setForm((s) => ({ ...s, format: e.target.value }))}
                >
                  <option value="csv">CSV</option>
                  <option value="json">JSON</option>
                  <option value="mp4">MP4</option>
                  <option value="txt">TXT</option>
                  <option value="parquet">Parquet</option>
                </select>
              </label>

              <label className="ui-input">
                <span className="ui-input__label">Публичность</span>
                <select
                    className="ui-input__field"
                    value={form.isPublic ? 'public' : 'private'}
                    onChange={(e) => setForm((s) => ({ ...s, isPublic: e.target.value === 'public' }))}
                >
                  <option value="public">Публичный</option>
                  <option value="private">Приватный</option>
                </select>
              </label>

              <FileDropInput
                  label="Загрузка"
                  file={form.file}
                  onChange={(file) => setForm((s) => ({ ...s, file }))}
              />

              <div className="create-ds__textarea">
                <TextArea
                    label="Описание"
                    placeholder="Enter dataset description..."
                    value={form.description}
                    onChange={(e) => setForm((s) => ({ ...s, description: e.target.value }))}
                    rows={6}
                />
              </div>

              <Button type="submit" variant="primary" fullWidth disabled={!canSubmit}>
                Создать
              </Button>

              {status.kind === 'ok' && <div className="create-ds__ok">{status.text}</div>}
              {status.kind === 'err' && <div className="create-ds__err">{status.text}</div>}
            </div>
          </form>
        </Card>
      </div>
  );
};

export default UploadPage;
