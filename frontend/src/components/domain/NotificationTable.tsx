import React from 'react';
import { NotificationItem } from '../../api/types';
import { Table, SortOrder } from '../ui/Table';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';

type SortState = { key: string; order: SortOrder };

type Props = {
    items: NotificationItem[];
    sort: SortState;
    onSortChange: (s: SortState) => void;
    onMarkRead: (id: number) => void;
};

export const NotificationsTable: React.FC<Props> = ({ items, sort, onSortChange, onMarkRead }) => {
    return (
        <Table<NotificationItem>
            rowKey={(n) => n.id}
            rows={items}
            sort={sort}
            onSortChange={onSortChange}
            columns={[
                { key: 'created_at', title: 'Date', sortable: true, render: (n) => n.created_at?.slice(0, 10) ?? '—' },
                {
                    key: 'is_read',
                    title: 'Status',
                    sortable: true,
                    render: (n) => (
                        <Badge tone={n.is_read ? 'info' : 'warning'}>
                            {n.is_read ? 'Read' : 'New'}
                        </Badge>
                    )
                },
                { key: 'message', title: 'Message', render: (n) => n.message },
                {
                    key: 'action',
                    title: '',
                    render: (n) => (
                        <Button variant="outline" size="sm" disabled={n.is_read} onClick={() => onMarkRead(n.id)}>
                            Mark read
                        </Button>
                    )
                }
            ]}
        />
    );
};
