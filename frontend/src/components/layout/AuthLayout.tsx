import React from 'react';

export const AuthLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    return (
        <div className="auth-layout">
            <div className="auth-layout__content">
                <div className="auth-layout__brand">
                    Dataset<span>Hub</span>
                </div>
                {children}
            </div>
        </div>
    );
};
