import React, { useId, useRef, useState } from 'react';

type Props = {
    label: string;
    file: File | null;
    onChange: (file: File | null) => void;
    accept?: string;
    hint?: string;
    error?: string;
};

export const FileDropInput: React.FC<Props> = ({ label, file, onChange, accept, hint, error }) => {
    const id = useId();
    const inputRef = useRef<HTMLInputElement | null>(null);
    const [dragOver, setDragOver] = useState(false);

    const pick = () => inputRef.current?.click();

    const onFiles = (files: FileList | null) => {
        const f = files?.[0] ?? null;
        onChange(f);
    };

    return (
        <div className={['ui-file', error ? 'ui-file--error' : ''].filter(Boolean).join(' ')}>
            <div
                className={[
                    'ui-file__control',
                    dragOver ? 'ui-file__control--drag' : ''
                ].filter(Boolean).join(' ')}
                onDragEnter={(e) => { e.preventDefault(); setDragOver(true); }}
                onDragOver={(e) => { e.preventDefault(); setDragOver(true); e.dataTransfer.dropEffect = 'copy'; }}
                onDragLeave={(e) => { e.preventDefault(); setDragOver(false); }}
                onDrop={(e) => {
                    e.preventDefault();
                    setDragOver(false);
                    onFiles(e.dataTransfer.files);
                }}
            >
                <div className="ui-file__label">{label}</div>

                <button
                    type="button"
                    className="ui-file__zone"
                    onClick={pick}
                    onDragEnter={(e) => { e.preventDefault(); setDragOver(true); }}
                    onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
                    onDragLeave={(e) => { e.preventDefault(); setDragOver(false); }}
                    onDrop={(e) => {
                        e.preventDefault();
                        setDragOver(false);
                        onFiles(e.dataTransfer.files);
                    }}
                >
          <span className="ui-file__zoneText">
            {file ? file.name : 'Drug here or click upload'}
          </span>
                </button>

                <button type="button" className="ui-file__icon" onClick={pick} aria-label="Upload">
                    ⬆
                </button>

                <input
                    id={id}
                    ref={inputRef}
                    type="file"
                    accept={accept}
                    className="ui-file__input"
                    onChange={(e) => onFiles(e.target.files)}
                />
            </div>

            {hint && !error && <div className="ui-file__hint">{hint}</div>}
            {error && <div className="ui-file__error">{error}</div>}
        </div>
    );
};
