import React from 'react';
import { useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Card } from '../components/ui/Card';
import { Pagination } from '../components/ui/Pagination';
import { Table, SortState } from '../components/ui/Table';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { useServices } from '../context/ServiceContext';
import { NotificationItem } from '../api/types';

const num = (sp: URLSearchParams, k: string, d: number) => Math.max(1, Number(sp.get(k) || d));
const get = (sp: URLSearchParams, k: string) => sp.get(k) || '';

export default function NotificationsPage() {
  const { notificationService } = useServices();
  const qc = useQueryClient();
  const [sp, setSp] = useSearchParams();

  const page = num(sp, 'page', 1);
  const sortKey = get(sp, 'sort') || 'created_at';
  const order = (get(sp, 'order') as SortState['order']) || 'desc';

  const set = (patch: Record<string, string | number | null>) => {
    const next = new URLSearchParams(sp);
    for (const [k, v] of Object.entries(patch)) {
      if (v === null || v === '') next.delete(k);
      else next.set(k, String(v));
    }
    setSp(next, { replace: true });
  };

  const query = useQuery({
    queryKey: ['notifications', page, sortKey, order],
    queryFn: async () => {
      const res = await notificationService.list(page, 10);
      const items = [...res.items];
      items.sort((a, b) => {
        const av = (a as any)[sortKey] ?? '';
        const bv = (b as any)[sortKey] ?? '';
        const res2 = String(av).localeCompare(String(bv));
        return order === 'desc' ? -res2 : res2;
      });
      return { ...res, items };
    }
  });

  const markRead = useMutation({
    mutationFn: (id: number) => notificationService.markRead(id, true),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] })
  });

  const totalPages = query.data ? Math.max(1, Math.ceil(query.data.meta.total / query.data.meta.limit)) : 1;

  return (
      <>
        <Card title="Notifications" subtitle="URL содержит page/sort/order">
          <Table<NotificationItem>
              rowKey={(n) => n.id}
              rows={query.data?.items ?? []}
              gridTemplate="140px 120px 1fr 160px"
              sort={{ key: sortKey, order }}
              onSortChange={(s) => set({ sort: s.key, order: s.order ?? null, page: 1 })}
              columns={[
                { key: 'created_at', title: 'Date', sortable: true, render: (n) => n.created_at?.slice(0, 10) ?? '—' },
                {
                  key: 'is_read',
                  title: 'Status',
                  sortable: true,
                  render: (n) => <Badge tone={n.is_read ? 'info' : 'warning'}>{n.is_read ? 'Read' : 'New'}</Badge>
                },
                { key: 'message', title: 'Message', sortable: false, render: (n) => n.message },
                {
                  key: 'action',
                  title: '',
                  sortable: false,
                  render: (n) => (
                      <Button
                          variant="outline"
                          size="sm"
                          disabled={n.is_read || markRead.isLoading}
                          onClick={() => markRead.mutate(n.id)}
                      >
                        Mark read
                      </Button>
                  )
                }
              ]}
          />
        </Card>

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
