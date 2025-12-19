import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { Pagination } from '../Pagination';

describe('Pagination', () => {
  it('renders pages and triggers onPageChange', () => {
    const onChange = vi.fn();
    render(<Pagination page={3} totalPages={7} onPageChange={onChange} />);

    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    expect(screen.getAllByText('…').length).toBeGreaterThan(0);

    fireEvent.click(screen.getByText('4'));
    expect(onChange).toHaveBeenCalledWith(4);
  });
});
