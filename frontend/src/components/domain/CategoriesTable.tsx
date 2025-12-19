import React from 'react';
import { Category } from '../../api/types';
import { Table, SortOrder } from '../ui/Table';

type SortState = { key: string; order: SortOrder };

type Props = {
    categories: Category[];
    sort: SortState;
    onSortChange: (s: SortState) => void;
};

export const CategoriesTable: React.FC<Props> = ({ categories, sort, onSortChange }) => {
    return (
        <Table<Category>
            rowKey={(c) => c.id}
            rows={categories}
            sort={sort}
            onSortChange={onSortChange}
            columns={[
                { key: 'name', title: 'Category', sortable: true, render: (c) => <strong>{c.name}</strong> },
                { key: 'description', title: 'Description', render: (c) => c.description ?? '—' }
            ]}
        />
    );
};
