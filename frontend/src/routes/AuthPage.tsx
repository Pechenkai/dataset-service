import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthForm } from '../components/domain/AuthForm';
import { useAuth } from '../context/AuthContext';

const AuthPage: React.FC = () => {
    const { login, registerAndLogin } = useAuth();
    const navigate = useNavigate();
    const [mode, setMode] = useState<'login' | 'register'>('login');

    return (
        <AuthForm
            mode={mode}
            onSwitchMode={() => setMode(mode === 'login' ? 'register' : 'login')}
            onSubmit={async ({ email, password, username, country }) => {
                if (mode === 'register') {
                    await registerAndLogin({ username: username || email, email, password, country: country || 'Unknown' });
                } else {
                    await login({ email, password });
                }
                navigate('/');
            }}
        />
    );
};

export default AuthPage;
