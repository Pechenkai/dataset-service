import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card } from '../components/ui/Card';
import { StatTile } from '../components/ui/StatTile';
import { useAuth } from '../context/AuthContext';
import { useServices } from '../context/ServiceContext';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';

const ProfilePage: React.FC = () => {
  const { session, logout } = useAuth();
  const { userService } = useServices();

  const userId = session?.user.id;
  const profileQuery = useQuery({
    queryKey: ['profile', userId],
    queryFn: () => userService.getUser(userId!),
    enabled: Boolean(userId)
  });

  if (!session) {
    return <Card title="Требуется вход">Войдите, чтобы увидеть профиль.</Card>;
  }

  if (profileQuery.isLoading) {
    return <Card title="Профиль">Загружаем…</Card>;
  }

  const user = profileQuery.data ?? session.user;

  return (
    <>
      <section className="page-hero">
        <h2>Профиль</h2>
        <p>Основная информация о пользователе.</p>
      </section>

      <Card
        title={user.username}
        subtitle={user.email}
        toolbar={
          <>
            <Badge tone={user.role === 'admin' ? 'success' : 'info'}>{user.role}</Badge>
            {user.is_blocked && <Badge tone="danger">Blocked</Badge>}
          </>
        }
        footer={<Button variant="ghost" onClick={logout}>Выйти</Button>}
      >
        <div className="ds-card__stats">
          <StatTile label="ID" value={`#${user.id}`} />
          <StatTile label="Страна" value={user.country || '—'} />
          <StatTile label="Дата регистрации" value={user.registration_date?.slice(0, 10) || '—'} />
          <StatTile label="Статус" value={user.is_blocked ? 'Заблокирован' : 'Активен'} />
          <StatTile label="Роль" value={user.role} />
        </div>
      </Card>
    </>
  );
};

export default ProfilePage;
