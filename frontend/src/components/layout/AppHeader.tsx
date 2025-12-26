import React from 'react';
import { NavLink } from 'react-router-dom';
import { Button } from '../ui/Button';
import { useAuth } from '../../context/AuthContext';
import { useTelegram } from '../../telegram/TelegramProvider';

export const AppHeader: React.FC = () => {
    const { isAuthenticated, session } = useAuth();
    const { isTelegram, webApp } = useTelegram();
    const isAdmin = session?.user.role === 'admin';

    return (
        <header className="app-header">
            <div className="app-header__inner">
                <div className="app-header__brand">Dataset<span>Hub</span></div>

                <nav className="app-header__nav">
                    <NavItem to="/categories">Categories</NavItem>
                    <NavItem to="/catalog">Datasets</NavItem>
                    <NavItem to="/notifications">Notifications</NavItem>
                    {isAuthenticated && <NavItem to="/my">My datasets</NavItem>}
                    {isAdmin && <NavItem to="/admin/users">Users</NavItem>}
                </nav>

                <div className="app-header__actions">
                    {isTelegram && (
                      <Button
                        variant="ghost"
                        onClick={() => webApp?.close?.() ?? webApp?.BackButton?.show()}
                        aria-label="Закрыть WebApp"
                      >
                        TG WebApp
                      </Button>
                    )}
                    {isAuthenticated ? (
                        <NavLink to="/profile">
                            <Button variant="ghost" aria-label="Profile">
                                👤
                            </Button>
                        </NavLink>
                    ) : (
                        <NavLink to="/auth">
                            <Button variant="outline">Login</Button>
                        </NavLink>
                    )}
                </div>
            </div>
        </header>
    );
};

const NavItem: React.FC<{ to: string; children: React.ReactNode }> = ({ to, children }) => (
    <NavLink
        to={to}
        className={({ isActive }) =>
            ['app-header__link', isActive ? 'app-header__link--active' : ''].join(' ')
        }
    >
        {children}
    </NavLink>
);
