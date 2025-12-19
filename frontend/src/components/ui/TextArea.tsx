import React from 'react';

type Props = React.TextareaHTMLAttributes<HTMLTextAreaElement> & {
    label?: string;
    hint?: string;
    error?: string;
    fullWidth?: boolean;
};

export const TextArea: React.FC<Props> = ({
                                              label,
                                              hint,
                                              error,
                                              fullWidth = true,
                                              className = '',
                                              rows = 5,
                                              ...props
                                          }) => {
    const cls = [
        'ui-textarea',
        fullWidth ? 'ui-textarea--full' : '',
        error ? 'ui-textarea--error' : '',
        className
    ]
        .filter(Boolean)
        .join(' ');

    return (
        <div className={cls}>
            <div className="ui-textarea__control">
                {label && <div className="ui-textarea__label">{label}</div>}
                <textarea className="ui-textarea__field" rows={rows} {...props} />
            </div>
            {hint && !error && <div className="ui-textarea__hint">{hint}</div>}
            {error && <div className="ui-textarea__error">{error}</div>}
        </div>
    );
};
