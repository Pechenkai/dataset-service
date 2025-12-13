import React from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { useServices } from '../context/ServiceContext';

const NotificationsPage: React.FC = () => {
  const { notificationService } = useServices();
  const queryClient = useQueryClient();
  const notificationsQuery = useQuery(['notifications'], () => notificationService.list());

  const markRead = useMutation((id: number) => notificationService.markRead(id, true), {
    onSuccess: () => queryClient.invalidateQueries(['notifications'])
  });

  return (
    <Card title="Уведомления" subtitle="API слой вынесен в notificationService">
      {notificationsQuery.data?.items.length === 0 && <p>Новых уведомлений нет.</p>}
      <div className="catalog-grid">
        {notificationsQuery.data?.items.map((note) => (
          <div key={note.id} className="stat-tile">
            <div className="meta-bar">
              <Badge tone={note.is_read ? 'info' : 'warning'}>
                {note.is_read ? 'Прочитано' : 'Новое'}
              </Badge>
              <span>{note.created_at?.slice(0, 10) || '—'}</span>
            </div>
            <p>{note.message}</p>
            <Button variant="ghost" onClick={() => markRead.mutate(note.id)} disabled={note.is_read}>
              Пометить прочитанным
            </Button>
          </div>
        ))}
      </div>
    </Card>
  );
};

export default NotificationsPage;
