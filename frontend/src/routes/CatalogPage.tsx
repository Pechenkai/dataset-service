import React, { useMemo, useState } from 'react';
import {Card} from '../components/ui/Card';
import {Input} from '../components/ui/Input';
import {Pagination} from '../components/ui/Pagination';
import {SortState, Table} from '../components/ui/Table';
import {Link} from 'react-router-dom';
import {useCatalogViewModel} from '../viewmodels/catalogViewModel';
import {formatSize} from '../utils/format';
import {Dataset} from '../api/types';
import {useQualityScores} from '../hooks/useQualityScores';
import {useDatasetWorker} from '../hooks/useDatasetWorker';
import {extractDatasetSize} from '../utils/dataset';

const CatalogPage: React.FC = () => {
    const {categories, datasets, filters, meta, isFetching, updateFilters} = useCatalogViewModel();
    const {scores, wasmReady} = useQualityScores(datasets);
    const workerStats = useDatasetWorker(datasets);
    const [sort, setSort] = useState<SortState | undefined>();
    const totalPages = meta ? Math.max(1, Math.ceil(meta.total / meta.limit)) : 1;
    const currentPage = filters.page ?? 1;

    const sortedDatasets = useMemo(() => {
        if (!sort || !sort.order) return datasets;
        const direction = sort.order === 'asc' ? 1 : -1;
        const next = [...datasets];
        next.sort((a, b) => {
            if (sort.key === 'quality') {
                const left = scores[a.id] ?? 0;
                const right = scores[b.id] ?? 0;
                if (left === right) return 0;
                return left > right ? direction : -direction;
            }
            if (sort.key === 'updated') {
                const left = a.updated_at ?? a.created_at ?? '';
                const right = b.updated_at ?? b.created_at ?? '';
                return left > right ? direction : -direction;
            }
            return 0;
        });
        return next;
    }, [datasets, scores, sort]);

    return (
        <>
            <section className="page-hero">
                <h2>Каталог датасетов</h2>
            </section>

            <Card title="Поиск и фильтры" subtitle="">
                <div className="filters-grid">
                    <Input
                        label="Поиск"
                        placeholder="Название или описание"
                        value={filters.search || ''}
                        onChange={(e) => updateFilters({search: e.target.value, page: 1})}
                    />
                    <label className="ui-input">
                        <span className="ui-input__label">Категория</span>
                        <select
                            className="ui-input__field"
                            value={filters.categoryId || ''}
                            onChange={(e) => updateFilters({
                                categoryId: e.target.value ? Number(e.target.value) : undefined,
                                page: 1
                            })}
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
                            onChange={(e) => updateFilters({visibility: e.target.value as any, page: 1})}
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
                        onChange={(e) => updateFilters({tags: e.target.value, page: 1})}
                        hint="Разделение запятыми"
                    />
                </div>
            </Card>

            {workerStats && (
                <Card title="Инсайты (Web Worker)" subtitle="Аналитика рассчитывается вне главного потока">
                    <div className="grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))' }}>
                        <div><strong>Всего:</strong> {workerStats.total}</div>
                        <div><strong>Публичных:</strong> {workerStats.publicCount}</div>
                        <div><strong>Приватных:</strong> {workerStats.privateCount}</div>
                        <div><strong>Средний рейтинг:</strong> {workerStats.averageRating || '—'}</div>
                        <div><strong>Средний размер:</strong> {workerStats.avgSize ? formatSize(workerStats.avgSize) : '—'}</div>
                        <div>
                            <strong>Топ теги:</strong>{' '}
                            {workerStats.topTags.length === 0
                                ? '—'
                                : workerStats.topTags.map((t) => `${t.tag} (${t.count})`).join(', ')}
                        </div>
                    </div>
                </Card>
            )}

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
                            rows={sortedDatasets}
                            sort={sort}
                            onSortChange={setSort}
                            gridTemplate="2fr 1.2fr 0.8fr 0.8fr 0.8fr 0.9fr 64px"
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
                                    key: 'quality',
                                    title: wasmReady ? 'Индекс качества (WASM)' : 'Индекс качества',
                                    sortable: true,
                                    render: (d) => {
                                        const score = scores[d.id];
                                        return typeof score === 'number' ? score : '—';
                                    }
                                },
                                {
                                    key: 'size',
                                    title: 'Размер',
                                    render: (d) => {
                                        const bytes = extractDatasetSize(d);
                                        return bytes ? formatSize(Number(bytes)) : '—';
                                    }
                                },
                                {
                                    key: 'updated',
                                    title: 'Обновлено',
                                    sortable: true,
                                    render: (d) => d.updated_at?.slice(0, 10) || d.created_at?.slice(0, 10) || '—'
                                },
                                {
                                    key: 'download', title: '', align: 'center', render: (d) => { /* см. ниже */
                                    }
                                },
                            ]}
                        />
                )}
            </Card>

            <div style={{marginTop: 16}}>
                <Pagination page={currentPage} totalPages={totalPages} onPageChange={(p) => updateFilters({page: p})}/>
            </div>
        </>
    );
};

export default CatalogPage;
