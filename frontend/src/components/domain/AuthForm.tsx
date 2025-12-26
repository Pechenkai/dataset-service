import React, { useState } from 'react';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { usePersistentState } from '../../hooks/usePersistentState';

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
  const [persisted, setPersisted] = usePersistentState('auth.form', {
    email: '',
    username: '',
    country: ''
  });
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await onSubmit({
        email: persisted.email,
        password,
        username: persisted.username,
        country: persisted.country
      });
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
            <Input
              label="Username"
              value={persisted.username}
              onChange={(e) => setPersisted((s) => ({ ...s, username: e.target.value }))}
              required
            />
            <Input
              label="Country"
              value={persisted.country}
              onChange={(e) => setPersisted((s) => ({ ...s, country: e.target.value }))}
              required
            />
          </>
        )}
        <Input
          label="Email"
          type="email"
          value={persisted.email}
          onChange={(e) => setPersisted((s) => ({ ...s, email: e.target.value }))}
          required
        />
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
