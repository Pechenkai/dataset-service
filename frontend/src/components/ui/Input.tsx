import React from 'react';
import clsx from 'clsx';

type InputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  hint?: string;
};

export const Input: React.FC<InputProps> = ({ label, hint, className, ...rest }) => {
  return (
    <label className="ui-input">
      {label && <span className="ui-input__label">{label}</span>}
      <input className={clsx('ui-input__field', className)} {...rest} />
      {hint && <span className="ui-input__hint">{hint}</span>}
    </label>
  );
};

export const inputStyles = `
.ui-input {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ui-input__label {
  color: var(--muted);
  font-size: 0.95rem;
}

.ui-input__field {
  width: 100%;
  padding: 12px 14px;
  border-radius: var(--radius-md);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: var(--surface);
  color: var(--text);
  outline: none;
  transition: border-color 150ms ease, box-shadow 150ms ease;
}

.ui-input__field:focus {
  border-color: var(--accent);
  box-shadow: 0 6px 18px rgba(125, 249, 194, 0.18);
}

.ui-input__hint {
  color: var(--muted);
  font-size: 0.85rem;
}
`;
