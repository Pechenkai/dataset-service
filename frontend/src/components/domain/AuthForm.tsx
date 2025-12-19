import React, { useState } from 'react';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';

type Mode = 'login' | 'register';

type Props = {
  mode: Mode;
  onSubmit: (payload: { email: string; password: string; username?: string; country?: string }) => Promise<void>;
  onSwitchMode: () => void;
  isSubmitting?: boolean;
};

export const AuthForm: React.FC<Props> = ({
  mode,
  onSubmit,
  onSwitchMode,
  isSubmitting = false
}) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [country, setCountry] = useState('');
  const [error, setError] = useState('');

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await onSubmit({ email, password, username, country });
    } catch {
      setError('Authentication failed');
    }
  };

  return (
    <form className="auth-form" onSubmit={submit}>
      <h2 className="auth-form__title">{mode === 'login' ? 'Sign in' : 'Register'}</h2>

      <div className="auth-form__fields">
        {mode === 'register' && (
          <>
            <Input label="Username" value={username} onChange={(e) => setUsername(e.target.value)} required />
            <Input label="Country" value={country} onChange={(e) => setCountry(e.target.value)} required />
          </>
        )}
        <Input label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <Input label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
      </div>

      <Button type="submit" variant="primary" disabled={isSubmitting}>
        {mode === 'login' ? 'Log in' : 'Create account'}
      </Button>

      {error && <div className="auth-form__error">{error}</div>}

      <div className="auth-form__switch">
        {mode === 'login' ? (
          <>
            Don’t have an account?
            <button type="button" onClick={onSwitchMode}>
              Register
            </button>
          </>
        ) : (
          <>
            Already have an account?
            <button type="button" onClick={onSwitchMode}>
              Sign in
            </button>
          </>
        )}
      </div>
    </form>
  );
};
