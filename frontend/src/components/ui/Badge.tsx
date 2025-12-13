import React from 'react';
import clsx from 'clsx';

type Tone = 'success' | 'warning' | 'danger' | 'info';

export const Badge: React.FC<{ tone?: Tone; children: React.ReactNode; className?: string }> = ({
  tone = 'info',
  children,
  className
}) => {
  return <span className={clsx('ui-badge', `ui-badge--${tone}`, className)}>{children}</span>;
};

export const badgeStyles = `
.ui-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 0.85rem;
  border: 1px solid transparent;
}

.ui-badge--info {
  background: rgba(107, 168, 255, 0.15);
  border-color: rgba(107, 168, 255, 0.35);
}

.ui-badge--success {
  background: rgba(126, 245, 183, 0.12);
  border-color: rgba(126, 245, 183, 0.35);
}

.ui-badge--warning {
  background: rgba(255, 209, 102, 0.12);
  border-color: rgba(255, 209, 102, 0.4);
}

.ui-badge--danger {
  background: rgba(255, 107, 107, 0.12);
  border-color: rgba(255, 107, 107, 0.35);
}
`;
