'use client';

import { usePathname } from 'next/navigation';
import { useEffect } from 'react';
import { composeTitle, setBaseTitle, titleKeyForPath } from '@/lib/documentTitle';
import { useT } from '@/lib/i18n/useT';

/** Keeps the tab title in step with the route and the language ("Sign in · WebChat"). */
export default function RouteTitle() {
  const pathname = usePathname();
  const { t } = useT();

  useEffect(() => {
    const key = titleKeyForPath(pathname);
    setBaseTitle(composeTitle(key ? t(key) : null, t('common.appName')));
  }, [pathname, t]);

  return null;
}
