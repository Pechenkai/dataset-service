import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { Tabs } from '../Tabs';

describe('Tabs', () => {
  it('renders items and changes active tab', () => {
    const onChange = vi.fn();
    render(
      <Tabs
        items={[
          { key: 'a', label: 'Tab A' },
          { key: 'b', label: 'Tab B' }
        ]}
        activeKey="a"
        onChange={onChange}
      />
    );

    fireEvent.click(screen.getByText('Tab B'));
    expect(onChange).toHaveBeenCalledWith('b');
  });
});
