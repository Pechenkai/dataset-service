import React from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Card } from '../components/ui/Card';
import { Table } from '../components/ui/Table';
import { Button } from '../components/ui/Button';
import { useAuth } from '../context/AuthContext';
import { useServices } from '../context/ServiceContext';
import { User } from '../api/types';

const gridTemplate = '100px 200px 240px 120px 260px';

const AdminUsersPage: React.FC = () => {
  const { session } = useAuth();
  const { userService } = useServices();
  const qc = useQueryClient();
  const isAdmin = session?.user.role === 'admin';

  const usersQuery = useQuery({
    queryKey: ['admin-users'],
    queryFn: () => userService.listUsers(),
    enabled: isAdmin
  });

  const toggleBlock = useMutation({
    mutationFn: (user: User) => userService.updateUser(user.id as number, { is_blocked: !user.is_blocked }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] })
  });

  const deleteUser = useMutation({
    mutationFn: (userId: number) => userService.deleteUser(userId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] })
  });

  if (!isAdmin) {
    return <Card title="Нет доступа">Только администратор может управлять пользователями.</Card>;
  }

  const rows = usersQuery.data?.items ?? [];

  return (
    <>
      <section className="page-hero">
        <h2>Управление пользователями</h2>
        <p>Блокировка, удаление, роли.</p>
      </section>

      <Card title="Пользователи" subtitle="Админ-доступ">
        <Table<User>
          rowKey={(u) => u.id}
          rows={rows}
          gridTemplate={gridTemplate}
          columns={[
            { key: 'id', title: 'ID', render: (u) => `#${u.id}` },
            { key: 'username', title: 'Имя', render: (u) => u.username },
            { key: 'email', title: 'Email', render: (u) => u.email },
            { key: 'role', title: 'Роль', render: (u) => u.role },
            {
              key: 'actions',
              title: 'Действия',
              render: (u) => (
                <div className="table-actions">
                  <Button
                    size="sm"
                    variant={u.is_blocked ? 'outline' : 'danger'}
                    onClick={() => toggleBlock.mutate(u)}
                  >
                    {u.is_blocked ? 'Разблокировать' : 'Заблокировать'}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => deleteUser.mutate(u.id as number)}
                    disabled={deleteUser.isLoading}
                  >
                    Удалить
                  </Button>
                </div>
              )
            }
          ]}
        />
      </Card>
    </>
  );
};

export default AdminUsersPage;
