import React from 'react';
import { Link } from 'react-router-dom';
import { Card } from '../components/ui/Card';

const NotFoundPage: React.FC = () => (
  <Card title="Страница не найдена">
    <p>Маршрут отсутствует или был удалён. Вернитесь в каталог.</p>
    <Link className="ui-button" to="/">
      К каталогу
    </Link>
  </Card>
);

export default NotFoundPage;
