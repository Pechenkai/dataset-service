import React, { useMemo } from 'react';
import { Link } from 'react-router-dom';
import { Category, Dataset } from '../../api/types';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { RatingChip } from '../ui/RatingChip';
import { StatTile } from '../ui/StatTile';
import { TagList } from '../ui/TagList';

type Props = {
    dataset: Dataset;
    category?: Category;
    onDownload?: (dataset: Dataset) => void;
};

const formatSize = (raw?: number | string | null) => {
    if (raw === undefined || raw === null) return '—';
    const bytes = typeof raw === 'string' ? Number(raw) : raw;
    if (!Number.isFinite(bytes) || bytes <= 0) return '—';
    const mb = bytes / 1024 / 1024;
    if (mb < 1024) return `${mb.toFixed(1)} MB`;
    return `${(mb / 1024).toFixed(1)} GB`;
};

const formatDate = (iso?: string) => (iso ? iso.slice(0, 10) : '—');

const joinTags = (tags?: string[]) => (tags && tags.length ? tags : []);

export const DatasetCard: React.FC<Props> = ({ dataset, category, onDownload }) => {
    const tags = useMemo(() => joinTags(dataset.metadata?.tags), [dataset.metadata?.tags]);

    const ratingLabel = dataset.rating_summary
        ? `${dataset.rating_summary.average.toFixed(1)}`
        : 'Rating';

    const versionValue = dataset.latest_version?.number || '—';

    const updatedValue = formatDate(dataset.updated_at || dataset.latest_version?.upload_date);

    const sizeValue = formatSize(
        dataset.metadata?.size ??
        dataset.latest_version?.metadata?.size ??
        (dataset.latest_version as any)?.size
    );

    return (
        <Card
            className="ds-card"
            title={<span className="ds-card__title">{dataset.name}</span>}
            subtitle={category ? category.name : 'Без категории'}
            toolbar={
                <Badge tone={dataset.is_public ? 'success' : 'warning'}>
                    {dataset.is_public ? 'Публичный' : 'Приватный'}
                </Badge>
            }
            footer={
                onDownload && (
                    <div className="ds-card__footer">
                        <Button fullWidth onClick={() => onDownload(dataset)}>
                            Download
                        </Button>
                    </div>
                )
            }
        >
            <div className="ds-card__layout">
                <div className="ds-card__left">
                    <div className="ds-card__rating">
                        <RatingChip label={ratingLabel} />
                    </div>
                    <TagList tags={tags.map((t) => t)} />
                    <div className="ds-card__stats">
                        <StatTile label="Size" value={sizeValue} />
                        <StatTile label="Updated" value={updatedValue} />
                        <StatTile label="Version" value={versionValue} />
                    </div>
                </div>
                <div className="ds-card__descBox">
                    <div className="ds-card__descTitle">Описание</div>
                    <div className="ds-card__desc">
                        {dataset.description || 'Описание пока пустое — самое время его добавить.'}
                    </div>
                </div>
            </div>
        </Card>
    );
};
