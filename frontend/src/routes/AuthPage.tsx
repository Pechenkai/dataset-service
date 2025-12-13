import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { useAuth } from '../context/AuthContext';

const AuthPage: React.FC = () => {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ email: '', password: '' });
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await login({ email: form.email, password: form.password });
      navigate('/');
    } catch (err) {
      setError('Не удалось авторизоваться. Проверьте данные.');
      console.error(err);
    }
  };

  return (
    <Card title="Вход" subtitle="Валидацию и вызов /auth/tokens берёт на себя AuthService">
      <form className="form-grid" onSubmit={handleSubmit}>
        <Input
          label="Email"
          type="email"
          value={form.email}
          onChange={(e) => setForm({ ...form, email: e.target.value })}
          required
        />
        <Input
          label="Пароль"
          type="password"
          value={form.password}
          onChange={(e) => setForm({ ...form, password: e.target.value })}
          required
        />
        <Button type="submit">Войти</Button>
        {error && <p style={{ color: 'salmon' }}>{error}</p>}
      </form>
    </Card>
  );
};

export default AuthPage;
