import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { TelegramThemeParams, TelegramWebApp, getTelegramWebApp } from './types';

type TelegramContextValue = {
  isTelegram: boolean;
  webApp: TelegramWebApp | null;
  theme: { colorScheme: 'light' | 'dark' | 'unknown'; params: TelegramThemeParams };
};

const defaultTheme = { colorScheme: 'light' as const, params: {} as TelegramThemeParams };

const TelegramContext = createContext<TelegramContextValue>({
  isTelegram: false,
  webApp: null,
  theme: defaultTheme
});

const applyThemeToDocument = (params: TelegramThemeParams, colorScheme: 'light' | 'dark' | 'unknown') => {
  const root = document.documentElement;
  if (params.bg_color) root.style.setProperty('--color-bg', params.bg_color);
  if (params.text_color) root.style.setProperty('--color-text-primary', params.text_color);
  if (params.hint_color) root.style.setProperty('--color-text-muted', params.hint_color);
  if (params.button_color) {
    root.style.setProperty('--color-primary', params.button_color);
    root.style.setProperty('--color-primary-light', params.button_color);
  }
  if (params.link_color) root.style.setProperty('--color-accent-1', params.link_color);
  root.dataset.tgColorScheme = colorScheme;
};

export const TelegramProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [webApp, setWebApp] = useState<TelegramWebApp | null>(() => getTelegramWebApp());
  const [theme, setTheme] = useState(defaultTheme);

  useEffect(() => {
    const tg = getTelegramWebApp();
    if (!tg) return;
    setWebApp(tg);
    tg.ready();
    tg.expand();
    tg.BackButton?.hide();

    const syncTheme = () => {
      const colorScheme = tg.colorScheme ?? 'light';
      const params = tg.themeParams ?? {};
      setTheme({ colorScheme, params });
      applyThemeToDocument(params, colorScheme);
    };

    syncTheme();
    tg.onEvent?.('themeChanged', syncTheme);
    return () => {
      tg.offEvent?.('themeChanged', syncTheme);
    };
  }, []);

  const value = useMemo<TelegramContextValue>(
    () => ({
      isTelegram: Boolean(webApp),
      webApp,
      theme
    }),
    [webApp, theme]
  );

  return <TelegramContext.Provider value={value}>{children}</TelegramContext.Provider>;
};

export const useTelegram = () => useContext(TelegramContext);
