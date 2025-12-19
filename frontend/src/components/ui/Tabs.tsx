import React from 'react';

export type TabItem = {
    key: string;
    label: string;
};

type Props = {
    items: TabItem[];
    activeKey: string;
    onChange: (key: string) => void;
};

export const Tabs: React.FC<Props> = ({ items, activeKey, onChange }) => {
    return (
        <div className="ui-tabs">
            {items.map((t) => (
                <button
                    key={t.key}
                    type="button"
                    className={['ui-tab', t.key === activeKey ? 'ui-tab--active' : ''].filter(Boolean).join(' ')}
                    onClick={() => onChange(t.key)}
                >
                    {t.label}
                </button>
            ))}
        </div>
    );
};
