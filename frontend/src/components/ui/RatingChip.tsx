import React from 'react';

type Props = {
    label: string;              // "Rating" или "4.8"
    icon?: React.ReactNode;     // по умолчанию звезда
    className?: string;
};

export const RatingChip: React.FC<Props> = ({ label, icon = '★', className = '' }) => {
    return (
        <div className={['ui-rating', className].filter(Boolean).join(' ')}>
            <div className="ui-rating__label">{label}</div>
            <div className="ui-rating__icon" aria-hidden>
                {icon}
            </div>
        </div>
    );
};
