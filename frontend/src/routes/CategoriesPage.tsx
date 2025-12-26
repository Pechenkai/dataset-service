import React, { useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Card } from '../components/ui/Card';
import { Pagination } from '../components/ui/Pagination';
import { Table, SortState } from '../components/ui/Table';
import { Input } from '../components/ui/Input';
import { Button } from '../components/ui/Button';
import { useServices } from '../context/ServiceContext';
import { useAuth } from '../context/AuthContext';
import { Category } from '../api/types';

const PAGE_SIZE = 10;

const get = (sp: URLSearchParams, k: string) => sp.get(k) || '';
const num = (sp: URLSearchParams, k: string, d: number) => Math.max(1, Number(sp.get(k) || d));

export default function CategoriesPage() {
    const { datasetService } = useServices();
    const { session } = useAuth();
    const qc = useQueryClient();
    const isAdmin = session?.user.role === 'admin';
    const [sp, setSp] = useSearchParams();
    const [form, setForm] = useState({ name: '', description: '' });

    const page = num(sp, 'page', 1);
    const sortKey = get(sp, 'sort') || 'name';
    const order = (get(sp, 'order') as SortState['order']) || 'asc';

    const set = (patch: Record<string, string | number | null>) => {
        const next = new URLSearchParams(sp);
        for (const [k, v] of Object.entries(patch)) {
            if (v === null || v === '') next.delete(k);
            else next.set(k, String(v));
        }
        setSp(next, { replace: true });
    };

    const query = useQuery({
        queryKey: ['categories', page],
        queryFn: () => datasetService.listCategories() // эндпоинт у тебя без пагинации; сорт/пейдж делаем на фронте
    });

    const createMutation = useMutation({
        mutationFn: () => datasetService.createCategory({ name: form.name.trim(), description: form.description.trim() || undefined }),
        onSuccess: (cat) => {
            qc.setQueryData(['categories', page], (prev: any) => {
                if (!prev?.items) return prev;
                return { ...prev, items: [...prev.items, cat] };
            });
            setForm({ name: '', description: '' });
        }
    });

    const all = query.data?.items ?? [];

    const sorted = useMemo(() => {
        const copy = [...all];
        copy.sort((a, b) => {
            const av = (a as any)[sortKey] ?? '';
            const bv = (b as any)[sortKey] ?? '';
            const res = String(av).localeCompare(String(bv));
            return order === 'desc' ? -res : res;
        });
        return copy;
    }, [all, sortKey, order]);

    const totalPages = Math.max(1, Math.ceil(sorted.length / PAGE_SIZE));
    const pageItems = sorted.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

    return (
        <>
            <Card title="Categories" subtitle="">
                <Table<Category>
                    rowKey={(c) => c.id}
                    rows={pageItems}
                    gridTemplate="240px 1fr"
                    sort={{ key: sortKey, order }}
                    onSortChange={(s) => set({ sort: s.key, order: s.order ?? null, page: 1 })}
                    columns={[
                        {
                            key: 'name',
                            title: 'Category',
                            sortable: true,
                            render: (c) => (
                                <Link to={`/catalog?category=${c.id}`} className="app-link">
                                    <strong>{c.name}</strong>
                                </Link>
                            )
                        },
                        { key: 'description', title: 'Description', sortable: false, render: (c) => c.description ?? '—' }
                    ]}
                />
            </Card>

            {isAdmin && (
                <Card title="Создать категорию" subtitle="" >
                    <div className="filters-grid">
                        <Input
                            label="Название"
                            value={form.name}
                            onChange={(e) => setForm({ ...form, name: e.target.value })}
                            required
                        />
                        <Input
                            label="Описание"
                            value={form.description}
                            onChange={(e) => setForm({ ...form, description: e.target.value })}
                        />
                    </div>
                    <div style={{ marginTop: 12 }}>
                        <Button
                            onClick={() => createMutation.mutate()}
                            disabled={!form.name.trim() || createMutation.isLoading}
                        >
                            Добавить категорию
                        </Button>
                    </div>
                </Card>
            )}

            <div style={{ marginTop: 16 }}>
                <Pagination
                    page={page}
                    totalPages={totalPages}
                    onPageChange={(p) => set({ page: p })}
                />
            </div>
        </>
    );
}
