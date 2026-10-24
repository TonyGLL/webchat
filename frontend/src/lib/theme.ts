export type Theme = 'light' | 'dark';

export const THEME_KEY = 'theme';

/** Stored choice wins; with none, follow the operating system. */
export function resolveTheme(stored: string | null, systemPrefersDark: boolean): Theme {
  if (stored === 'light' || stored === 'dark') return stored;
  return systemPrefersDark ? 'dark' : 'light';
}

/**
 * Runs before first paint (inlined in <head>) so the page never flashes the wrong theme.
 * Kept as a string because it executes outside React.
 */
export const themeInitScript =
  `(function(){try{var s=localStorage.getItem('${THEME_KEY}');var d=s==='dark'||(s!=='light'&&matchMedia('(prefers-color-scheme: dark)').matches);document.documentElement.classList.toggle('dark',d)}catch(e){}` +
  // Same rule as resolveLocale (lib/i18n): stored choice, else first supported browser language, else English.
  `try{var l=null;try{l=localStorage.getItem('locale')}catch(e){}var v=['en','es'];if(v.indexOf(l)<0){l='en';var n=navigator.languages||[navigator.language];for(var i=0;i<n.length;i++){var b=String(n[i]).toLowerCase().split('-')[0];if(v.indexOf(b)>=0){l=b;break}}}document.documentElement.lang=l}catch(e){}})()`;

export function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle('dark', theme === 'dark');
  try {
    localStorage.setItem(THEME_KEY, theme);
  } catch {
    // storage blocked: the choice just won't persist
  }
}
