import { render, screen } from '@testing-library/react';
import React from 'react';
import { TextArea } from '../TextArea';

describe('TextArea', () => {
  it('renders label, hint and error', () => {
    const { rerender } = render(<TextArea label="Note" hint="Optional" placeholder="comment" />);
    expect(screen.getByText('Note')).toBeInTheDocument();
    expect(screen.getByText('Optional')).toBeInTheDocument();

    rerender(<TextArea label="Note" error="Required" />);
    expect(screen.getByText('Required')).toBeInTheDocument();
  });
});
