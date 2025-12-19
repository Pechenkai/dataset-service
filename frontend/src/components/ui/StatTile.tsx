import React from 'react';

type Props = {
    label: string;
    value: React.ReactNode;
    className?: string;
};

export const StatTile: React.FC<Props> = ({ label, value, className = '' }) => {
    return (
        <div className={['ui-stat', className].filter(Boolean).join(' ')}>
            <div className="ui-stat__label">{label}</div>
            <div className="ui-stat__value">{value}</div>
        </div>
    );
};
