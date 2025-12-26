export type TelegramThemeParams = {
  bg_color?: string;
  text_color?: string;
  hint_color?: string;
  link_color?: string;
  button_color?: string;
  button_text_color?: string;
};

export type TelegramWebApp = {
  ready: () => void;
  expand: () => void;
  close?: () => void;
  colorScheme?: 'light' | 'dark' | 'unknown';
  themeParams?: TelegramThemeParams;
  onEvent?: (event: 'themeChanged', handler: () => void) => void;
  offEvent?: (event: 'themeChanged', handler: () => void) => void;
  BackButton?: {
    isVisible: boolean;
    show: () => void;
    hide: () => void;
    onClick: (handler: () => void) => void;
    offClick: (handler: () => void) => void;
  };
  MainButton?: {
    text: string;
    color?: string;
    textColor?: string;
    isVisible: boolean;
    setParams?: (params: { text?: string; color?: string; text_color?: string }) => void;
    show: () => void;
    hide: () => void;
  };
  HapticFeedback?: {
    impactOccurred: (style?: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft') => void;
  };
};

declare global {
  interface Window {
    Telegram?: {
      WebApp?: TelegramWebApp;
    };
  }
}

export const getTelegramWebApp = (): TelegramWebApp | null => {
  if (typeof window === 'undefined') return null;
  return window.Telegram?.WebApp ?? null;
};
