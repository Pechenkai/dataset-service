import React from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Card } from '../components/ui/Card';
import { Pagination } from '../components/ui/Pagination';
import { Table, SortState } from '../components/ui/Table';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { ReviewList } from '../components/domain/ReviewList';
import { ReviewForm } from '../components/domain/ReviewForm';
import { DatasetCard } from '../components/domain/DatasetCard';
import { useDatasetViewModel } from '../viewmodels/datasetViewModel';
import { useServices } from '../context/ServiceContext';

const num = (sp: URLSearchParams, k: string, d: number) => Math.max(1, Number(sp.get(k) || d));
const get = (sp: URLSearchParams, k: string) => sp.get(k) || '';

export default function DatasetPage() {
    const { datasetId } = useParams<{ datasetId: string }>();
    const id = Number(datasetId);
    const [sp, setSp] = useSearchParams();

    const vpage = num(sp, 'vpage', 1);
    const vsort = get(sp, 'vsort') || 'upload_date';
    const vorder = (get(sp, 'vorder') as SortState['order']) || 'desc';

    const rpage = num(sp, 'rpage', 1);

    const set = (patch: Record<string, string | number | null>) => {
        const next = new URLSearchParams(sp);
        for (const [k, v] of Object.entries(patch)) {
            if (v === null || v === '') next.delete(k);
            else next.set(k, String(v));
        }
        setSp(next, { replace: true });
    };

    const vm = useDatasetViewModel(id);
    const { datasetQuery, versionsQuery, reviewsQuery, createReview, subscribe } = vm;
    const { datasetService } = useServices();
    const categoriesQuery = useQuery({
        queryKey: ['categories'],
        queryFn: () => datasetService.listCategories()
    });
    const category = categoriesQuery.data?.items.find((c) => c.id === dataset.category_id);

    const [reviewError, setReviewError] = React.useState<string | null>(null);

    if (datasetQuery.isLoading) return <p>Загружаем датасет…</p>;
    if (!datasetQuery.data) return <p>Датасет не найден.</p>;

    const dataset = datasetQuery.data;

    const versions = versionsQuery.data?.items ?? [];
    const reviews = reviewsQuery.data?.items ?? [];

    const totalVersionsPages =
        versionsQuery.data ? Math.max(1, Math.ceil(versionsQuery.data.meta.total / versionsQuery.data.meta.limit)) : 1;

    const totalReviewsPages =
        reviewsQuery.data ? Math.max(1, Math.ceil(reviewsQuery.data.meta.total / reviewsQuery.data.meta.limit)) : 1;

    return (
        <>
            <DatasetCard dataset={dataset} category={category} />

            <Card title="Actions">
                <div className="ds-card__actions" style={{ gap: 12 }}>
                    <Button onClick={() => subscribe.mutate()}>Subscribe</Button>
                    <Button variant="outline" onClick={() => window.history.back()}>Назад</Button>
                </div>
            </Card>

            <Card title="Versions">
                <Table<any>
                    rowKey={(v) => v.id}
                    rows={[...versions].sort((a, b) => {
                        const av = (a as any)[vsort] ?? '';
                        const bv = (b as any)[vsort] ?? '';
                        const res = String(av).localeCompare(String(bv));
                        return vorder === 'desc' ? -res : res;
                    })}
                    gridTemplate="140px 140px 120px 1fr 56px"
                    sort={{ key: vsort, order: vorder }}
                    onSortChange={(s) => set({ vsort: s.key, vorder: s.order ?? null, vpage: 1 })}
                    columns={[
                        { key: 'number', title: 'Version', sortable: true, render: (v) => <strong>{v.number}</strong> },
                        { key: 'upload_date', title: 'Updated', sortable: true, render: (v) => v.upload_date?.slice(0, 10) ?? '—' },
                        {
                            key: 'size',
                            title: 'Size',
                            sortable: true,
                            align: 'right',
                            render: (v) => (v.metadata?.size ? `${Math.round(v.metadata.size / 1024 / 1024)} MB` : '—')
                        },
                        { key: 'change_log', title: 'Notes', sortable: false, render: (v) => v.change_log ?? '—' },
                        {
                            key: 'download',
                            title: '',
                            sortable: false,
                            render: (v) =>
                                v.file_url ? (
                                    <a className="ui-button ui-button--outline ui-button--sm" href={v.file_url} target="_blank" rel="noreferrer">
                                        ⬇
                                    </a>
                                ) : (
                                    <button className="ui-button ui-button--ghost ui-button--sm" disabled>⬇</button>
                                )
                        }
                    ]}
                />
                <div style={{ marginTop: 16 }}>
                    <Pagination page={vpage} totalPages={totalVersionsPages} onPageChange={(p) => set({ vpage: p })} />
                </div>
            </Card>

            <Card title="Отзывы">
                <div style={{ marginBottom: 12 }}>
                    <ReviewForm
                        isSubmitting={createReview.isLoading}
                        onSubmit={async ({ rating, text }) => {
                            setReviewError(null);
                            try {
                                await createReview.mutateAsync({ rating, text });
                            } catch (e: any) {
                                setReviewError(e?.message || 'Не удалось отправить отзыв.');
                            }
                        }}
                    />
                    {reviewError && <div className="create-ds__err">{reviewError}</div>}
                </div>
                <ReviewList reviews={reviews} />
                <div style={{ marginTop: 16 }}>
                    <Pagination page={rpage} totalPages={totalReviewsPages} onPageChange={(p) => set({ rpage: p })} />
                </div>
            </Card>
        </>
    );
}
