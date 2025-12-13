import React from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { Button } from '../ui/Button';

const links = [
  { to: '/', label: 'Каталог' },
  { to: '/upload', label: 'Публикация' },
  { to: '/notifications', label: 'Уведомления' }
];

export const AppShell: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { session, isAuthenticated, logout } = useAuth();

  return (
    <div className="app-shell">
      <aside className="app-shell__nav">
        <div className="app-shell__brand">
          <div>
            <h1>Dataset Platform</h1>
            <p className="meta-bar">SPA · v2 API</p>
          </div>
        </div>
        <nav className="app-shell__links">
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              className={({ isActive }) =>
                ['app-shell__link', isActive ? 'app-shell__link--active' : ''].filter(Boolean).join(' ')
              }
              end={link.to === '/'}
            >
              {link.label}
            </NavLink>
          ))}
          {isAuthenticated ? (
            <Button variant="ghost" onClick={logout}>
              Выйти ({session?.user.username})
            </Button>
          ) : (
            <NavLink
              to="/auth"
              className={({ isActive }) =>
                ['app-shell__link', isActive ? 'app-shell__link--active' : ''].filter(Boolean).join(' ')
              }
            >
              Войти
            </NavLink>
          )}
        </nav>
      </aside>
      <main className="app-shell__content">{children}</main>
    </div>
  );
};
