import { render, screen } from '@testing-library/react';
import React from 'react';
import { Input } from '../Input';

describe('Input', () => {
  it('shows label, hint and error', () => {
    const { rerender } = render(<Input label="Email" hint="optional" placeholder="e@x.y" />);
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('optional')).toBeInTheDocument();

    rerender(<Input label="Email" error="Required" />);
    expect(screen.getByText('Required')).toBeInTheDocument();
  });
});
