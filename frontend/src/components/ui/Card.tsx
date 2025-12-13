import React from 'react';
import clsx from 'clsx';

type CardProps = {
  title?: string;
  subtitle?: string;
  children: React.ReactNode;
  toolbar?: React.ReactNode;
  className?: string;
};

export const Card: React.FC<CardProps> = ({ title, subtitle, children, toolbar, className }) => {
  return (
    <section className={clsx('ui-card', className)}>
      {(title || toolbar) && (
        <header className="ui-card__header">
          <div>
            {title && <h3 className="ui-card__title">{title}</h3>}
            {subtitle && <p className="ui-card__subtitle">{subtitle}</p>}
          </div>
          {toolbar && <div className="ui-card__toolbar">{toolbar}</div>}
        </header>
      )}
      <div className="ui-card__body">{children}</div>
    </section>
  );
};

export const cardStyles = `
.ui-card {
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.06), rgba(255, 255, 255, 0.02));
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-lg);
  padding: 18px;
  box-shadow: var(--shadow);
}

.ui-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.ui-card__title {
  margin: 0;
  font-size: 1.1rem;
}

.ui-card__subtitle {
  margin: 4px 0 0;
  color: var(--muted);
}

.ui-card__body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ui-card__toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
}
`;
