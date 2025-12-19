import React from 'react';

export type SortOrder = 'asc' | 'desc' | null;

export type Column<T> = {
    key: keyof T | string;
    title: string;
    sortable?: boolean;
    align?: 'left' | 'right' | 'center';
    render?: (row: T) => React.ReactNode;
};

export type SortState = {
    key: string;
    order: SortOrder;
};

type Props<T> = {
    columns: Column<T>[];
    rows: T[];
    sort?: SortState;
    onSortChange?: (next: SortState) => void;
    rowKey: (row: T) => React.Key;
    gridTemplate?: string;
};

export function Table<T>({
                             columns,
                             rows,
                             sort,
                             onSortChange,
                             rowKey,
                             gridTemplate
                         }: Props<T>) {
    const toggleSort = (key: string) => {
        if (!onSortChange) return;

        if (sort?.key !== key) onSortChange({ key, order: 'asc' });
        else if (sort.order === 'asc') onSortChange({ key, order: 'desc' });
        else onSortChange({ key, order: null });
    };

    return (
        <div className="ui-table">
            <div className="ui-table__head" style={gridTemplate ? { gridTemplateColumns: gridTemplate } : undefined}>
                {columns.map((col) => {
                    const active = sort?.key === col.key;
                    return (
                        <div
                            key={String(col.key)}
                            className={[
                                'ui-table__cell',
                                'ui-table__cell--head',
                                col.sortable ? 'ui-table__cell--sortable' : '',
                                active ? 'ui-table__cell--active' : '',
                                col.align ? `ui-table__cell--${col.align}` : ''
                            ].filter(Boolean).join(' ')}
                            onClick={() => col.sortable && toggleSort(String(col.key))}
                        >
                            {col.title}
                            {col.sortable && (
                                <span className="ui-table__sort">
                  {active && sort?.order === 'asc' && '▲'}
                                    {active && sort?.order === 'desc' && '▼'}
                                    {!active && '↕'}
                </span>
                            )}
                        </div>
                    );
                })}
            </div>

            <div className="ui-table__body">
                {rows.map((row) => (
                    <div
                        key={rowKey(row)}
                        className="ui-table__row"
                        style={gridTemplate ? { gridTemplateColumns: gridTemplate } : undefined}
                    >
                        {columns.map((col) => (
                            <div
                                key={String(col.key)}
                                className={[
                                    'ui-table__cell',
                                    col.align ? `ui-table__cell--${col.align}` : ''
                                ].filter(Boolean).join(' ')}
                            >
                                {col.render ? col.render(row) : String((row as any)[col.key])}
                            </div>
                        ))}
                    </div>
                ))}
            </div>
        </div>
    );
}
