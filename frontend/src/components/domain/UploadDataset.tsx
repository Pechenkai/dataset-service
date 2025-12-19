import React from 'react';
import { Input } from '../ui/Input';
import { TextArea } from '../ui/TextArea';
import { Button } from '../ui/Button';
import { FileDropInput } from '../ui/FileDropInput';

export type UploadDatasetFormState = {
  name: string;
  tags: string;
  description: string;
  file: File | null;
};

export type UploadDatasetStatus = { kind: 'idle' | 'ok' | 'err'; text?: string };

type Props = {
  value: UploadDatasetFormState;
  onChange: <K extends keyof UploadDatasetFormState>(key: K, value: UploadDatasetFormState[K]) => void;
  onSubmit: (e: React.FormEvent) => void;
  canSubmit: boolean;
  status: UploadDatasetStatus;
};

export const UploadDataset: React.FC<Props> = ({ value, onChange, onSubmit, canSubmit, status }) => {
  return (
    <div className="create-ds">
      <div className="create-ds__hint">Create dataset</div>

      <div className="ui-card">
        <form className="create-ds__form" onSubmit={onSubmit}>
          <div className="create-ds__fields">
            <Input
              label="Название"
              placeholder="Input field"
              value={value.name}
              onChange={(e) => onChange('name', e.target.value)}
              required
            />

            <Input
              label="Теги"
              placeholder="Input field"
              value={value.tags}
              onChange={(e) => onChange('tags', e.target.value)}
              hint="Через запятую"
            />

            <FileDropInput label="Загрузка" file={value.file} onChange={(file) => onChange('file', file)} />

            <div className="create-ds__textarea">
              <TextArea
                label="Описание"
                placeholder="Enter dataset description..."
                value={value.description}
                onChange={(e) => onChange('description', e.target.value)}
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
      </div>
    </div>
  );
};
