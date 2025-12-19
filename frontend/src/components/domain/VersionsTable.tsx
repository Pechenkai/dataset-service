import React from 'react';
import { DatasetVersion } from '../../api/types';
import { Table, SortOrder } from '../ui/Table';
import { Button } from '../ui/Button';

type SortState = { key: string; order: SortOrder };

type Props = {
    versions: DatasetVersion[];
    sort: SortState;
    onSortChange: (s: SortState) => void;
};

export const VersionsTable: React.FC<Props> = ({ versions, sort, onSortChange }) => {
    return (
        <Table<DatasetVersion>
            rowKey={(v) => v.id}
            rows={versions}
            sort={sort}
            onSortChange={onSortChange}
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
                { key: 'change_log', title: 'Notes', render: (v) => v.change_log ?? '—' },
                {
                    key: 'download',
                    title: '',
                    render: (v) =>
                        v.file_url ? (
                            <a className="ui-button ui-button--outline ui-button--sm" href={v.file_url} target="_blank" rel="noreferrer">
                                ⬇
                            </a>
                        ) : (
                            <Button variant="ghost" size="sm" disabled>
                                ⬇
                            </Button>
                        )
                }
            ]}
        />
    );
};
