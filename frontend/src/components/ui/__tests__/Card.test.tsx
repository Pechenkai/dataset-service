import { render, screen } from '@testing-library/react';
import React from 'react';
import { Card } from '../Card';

describe('Card', () => {
  it('renders title, subtitle, toolbar and footer', () => {
    render(
      <Card title="Hello" subtitle="World" toolbar={<span>TB</span>} footer={<span>Footer</span>}>
        <div>Body</div>
      </Card>
    );

    expect(screen.getByText('Hello')).toBeInTheDocument();
    expect(screen.getByText('World')).toBeInTheDocument();
    expect(screen.getByText('TB')).toBeInTheDocument();
    expect(screen.getByText('Footer')).toBeInTheDocument();
    expect(screen.getByText('Body')).toBeInTheDocument();
  });
});
