import React from 'react';
import clsx from 'clsx';

type ButtonVariant = 'primary' | 'ghost' | 'outline';

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  icon?: React.ReactNode;
};

export const Button: React.FC<ButtonProps> = ({ variant = 'primary', icon, children, className, ...rest }) => {
  return (
    <button
      className={clsx(
        'ui-button',
        `ui-button--${variant}`,
        className
      )}
      {...rest}
    >
      {icon && <span className="ui-button__icon">{icon}</span>}
      <span>{children}</span>
    </button>
  );
};

export const buttonStyles = `
.ui-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  background: linear-gradient(120deg, rgba(125, 249, 194, 0.2), rgba(107, 168, 255, 0.14));
  color: var(--text);
  cursor: pointer;
  transition: transform 150ms ease, box-shadow 150ms ease, border-color 150ms ease;
  font-weight: 600;
}

.ui-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.ui-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.ui-button--ghost {
  background: transparent;
  border-color: rgba(255, 255, 255, 0.1);
}

.ui-button--outline {
  background: transparent;
  border-color: var(--accent);
}

.ui-button__icon {
  display: grid;
  place-items: center;
}
`;
