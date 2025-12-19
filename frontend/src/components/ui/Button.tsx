import React from 'react';

export type ButtonVariant = 'primary' | 'outline' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md';

type Props = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  fullWidth?: boolean;
};

export const Button: React.FC<Props> = ({
                                          variant = 'primary',
                                          size = 'md',
                                          fullWidth = false,
                                          className = '',
                                          ...props
                                        }) => {
  const cls = [
    'ui-button',
    `ui-button--${variant}`,
    `ui-button--${size}`,
    fullWidth ? 'ui-button--full' : '',
    className
  ]
      .filter(Boolean)
      .join(' ');

  return <button {...props} className={cls} />;
};
