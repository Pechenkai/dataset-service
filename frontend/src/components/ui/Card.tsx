import React from 'react';

type Props = {
    title?: React.ReactNode;
    subtitle?: React.ReactNode;
    toolbar?: React.ReactNode;
    footer?: React.ReactNode;
    children: React.ReactNode;
    className?: string;
};

export const Card: React.FC<Props> = ({ title, subtitle, toolbar, footer, children, className = '' }) => {
    return (
        <section className={['ui-card', className].filter(Boolean).join(' ')}>
            {(title || subtitle || toolbar) && (
                <header className="ui-card__header">
                    <div className="ui-card__headings">
                        {title && <div className="ui-card__title">{title}</div>}
                        {subtitle && <div className="ui-card__subtitle">{subtitle}</div>}
                    </div>
                    {toolbar && <div className="ui-card__toolbar">{toolbar}</div>}
                </header>
            )}

            <div className="ui-card__body">{children}</div>

            {footer && <footer className="ui-card__footer">{footer}</footer>}
        </section>
    );
};