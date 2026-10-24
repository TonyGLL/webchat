'use client';

import dynamic from 'next/dynamic';

/** The picker and its emoji dataset load on first open, so they stay out of the dashboard's first-load JS. */
const LazyEmojiPicker = dynamic(() => import('./EmojiPicker'), { ssr: false });

export default LazyEmojiPicker;
