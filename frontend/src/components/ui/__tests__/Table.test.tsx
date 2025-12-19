import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { Table } from '../Table';

type Row = { id: number; name: string; age: number };

describe('Table', () => {
  it('renders rows and toggles sort', () => {
    const rows: Row[] = [
      { id: 1, name: 'Alice', age: 30 },
      { id: 2, name: 'Bob', age: 20 }
    ];
    const onSortChange = vi.fn();

    render(
      <Table<Row>
        rowKey={(r) => r.id}
        rows={rows}
        sort={{ key: 'name', order: 'asc' }}
        onSortChange={onSortChange}
        columns={[
          { key: 'name', title: 'Name', sortable: true },
          { key: 'age', title: 'Age', sortable: true, align: 'right' },
          { key: 'custom', title: 'Custom', render: (r) => <span data-testid={`row-${r.id}`}>{r.name}</span> }
        ]}
      />
    );

    expect(screen.getAllByText('Alice').length).toBeGreaterThan(0);
    expect(screen.getByTestId('row-1')).toHaveTextContent('Alice');

    fireEvent.click(screen.getByText('Name'));
    expect(onSortChange).toHaveBeenCalledWith({ key: 'name', order: 'desc' });
  });
});
