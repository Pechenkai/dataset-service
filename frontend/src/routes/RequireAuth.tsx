import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const RequireAuth: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { isAuthenticated } = useAuth();
    const loc = useLocation();

    if (!isAuthenticated) {
        const next = encodeURIComponent(loc.pathname + loc.search);
        return <Navigate to={`/auth?next=${next}`} replace />;
    }
    return <>{children}</>;
};
