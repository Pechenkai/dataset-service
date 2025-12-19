import React, { useMemo } from 'react';

type Props = {
    page: number;          // 1-based
    totalPages: number;
    onPageChange: (page: number) => void;

    siblingCount?: number; // сколько страниц вокруг текущей показывать
    className?: string;
};

type Item =
    | { kind: 'page'; value: number; active: boolean }
    | { kind: 'dots'; key: string };

export const Pagination: React.FC<Props> = ({
                                                page,
                                                totalPages,
                                                onPageChange,
                                                siblingCount = 1,
                                                className = ''
                                            }) => {
    const items = useMemo<Item[]>(() => buildItems(page, totalPages, siblingCount), [page, totalPages, siblingCount]);

    if (totalPages <= 1) return null;

    return (
        <div className={['ui-pagination', className].filter(Boolean).join(' ')}>
            {items.map((it) => {
                if (it.kind === 'dots') {
                    return (
                        <button key={it.key} className="ui-page ui-page--dots" type="button" disabled>
                            …
                        </button>
                    );
                }

                return (
                    <button
                        key={it.value}
                        className={['ui-page', it.active ? 'ui-page--active' : ''].filter(Boolean).join(' ')}
                        type="button"
                        onClick={() => onPageChange(it.value)}
                        aria-current={it.active ? 'page' : undefined}
                    >
                        {it.value}
                    </button>
                );
            })}
        </div>
    );
};

function buildItems(page: number, total: number, siblingCount: number): Item[] {
    const safePage = clamp(page, 1, total);

    const first = 1;
    const last = total;

    const left = Math.max(first + 1, safePage - siblingCount);
    const right = Math.min(last - 1, safePage + siblingCount);

    const out: Item[] = [];
    out.push({ kind: 'page', value: first, active: safePage === first });

    if (left > first + 1) out.push({ kind: 'dots', key: 'dots-left' });

    for (let p = left; p <= right; p++) {
        out.push({ kind: 'page', value: p, active: p === safePage });
    }

    if (right < last - 1) out.push({ kind: 'dots', key: 'dots-right' });

    if (last !== first) out.push({ kind: 'page', value: last, active: safePage === last });

    return out;
}

function clamp(v: number, min: number, max: number) {
    return Math.max(min, Math.min(max, v));
}
