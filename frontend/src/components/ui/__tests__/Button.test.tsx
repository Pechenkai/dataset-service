import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { Button } from '../Button';

describe('Button', () => {
  it('renders variant, size and calls handler', () => {
    const onClick = vi.fn();
    render(
      <Button variant="outline" size="sm" fullWidth onClick={onClick}>
        Click me
      </Button>
    );

    const btn = screen.getByRole('button', { name: /click me/i });
    expect(btn.className).toContain('ui-button--outline');
    expect(btn.className).toContain('ui-button--sm');
    expect(btn.className).toContain('ui-button--full');

    fireEvent.click(btn);
    expect(onClick).toHaveBeenCalled();
  });
});
