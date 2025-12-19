import React from 'react';

type Props = Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> & {
  label?: string;
  hint?: string;
  error?: string;
  fullWidth?: boolean;
};

export const Input: React.FC<Props> = ({
                                         label,
                                         hint,
                                         error,
                                         fullWidth = true,
                                         className = '',
                                         ...props
                                       }) => {
  const cls = [
    'ui-input',
    fullWidth ? 'ui-input--full' : '',
    error ? 'ui-input--error' : '',
    className
  ]
      .filter(Boolean)
      .join(' ');

  return (
      <div className={cls}>
        <div className="ui-input__control">
          {label && <div className="ui-input__label">{label}</div>}
          <input className="ui-input__field" {...props} />
        </div>
        {hint && !error && <div className="ui-input__hint">{hint}</div>}
        {error && <div className="ui-input__error">{error}</div>}
      </div>
  );
};
