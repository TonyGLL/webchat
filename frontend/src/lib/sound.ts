let context: AudioContext | null = null;

function audio(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  try {
    const Ctor = window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!Ctor) return null;
    context ??= new Ctor();
    if (context.state === 'suspended') void context.resume();
    return context;
  } catch {
    return null;
  }
}

function beep(ctx: AudioContext, frequency: number, start: number, duration: number, volume: number) {
  const oscillator = ctx.createOscillator();
  const gain = ctx.createGain();
  oscillator.type = 'sine';
  oscillator.frequency.value = frequency;
  // A quick fade in and out avoids the click of a hard-edged tone.
  gain.gain.setValueAtTime(0.0001, start);
  gain.gain.exponentialRampToValueAtTime(volume, start + 0.015);
  gain.gain.exponentialRampToValueAtTime(0.0001, start + duration);
  oscillator.connect(gain).connect(ctx.destination);
  oscillator.start(start);
  oscillator.stop(start + duration + 0.02);
}

/**
 * Plays a short synthesized chime (no audio files to ship). Rooms get one soft note, direct
 * messages a rising pair so they are recognisable without looking. Silently does nothing if the
 * browser has not allowed audio yet (it needs a prior click or key press on the page).
 */
export function playMessageSound(kind: 'room' | 'direct') {
  const ctx = audio();
  if (!ctx || ctx.state !== 'running') return;
  const now = ctx.currentTime;
  if (kind === 'direct') {
    beep(ctx, 660, now, 0.16, 0.12);
    beep(ctx, 880, now + 0.11, 0.2, 0.12);
  } else {
    beep(ctx, 740, now, 0.2, 0.1);
  }
}
