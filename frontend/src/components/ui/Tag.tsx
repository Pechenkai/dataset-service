import React from 'react';

type Props = {
    children: React.ReactNode;
    tone?: 'default' | 'accent';
    className?: string;
};

export const Tag: React.FC<Props> = ({ children, tone = 'accent', className = '' }) => {
    return (
        <span className={['ui-tag', `ui-tag--${tone}`, className].filter(Boolean).join(' ')}>
      {children}
    </span>
    );
};
