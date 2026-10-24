import { cn } from '@/lib/utils';
import { initials } from '@/lib/initials';

const sizes = {
  xs: 'h-5 w-5 rounded text-[9px]',
  sm: 'h-7 w-7 rounded-md text-[11px]',
  md: 'h-9 w-9 rounded-lg text-xs',
} as const;

const tones = {
  /** The signed-in user. */
  me: 'bg-primary text-primary-foreground',
  /** Everyone else. */
  other: 'bg-secondary text-foreground',
  /** Admins and moderators stand out gently. */
  soft: 'bg-soft text-primary',
} as const;

export default function Avatar({
  name,
  size = 'md',
  tone = 'other',
  className,
}: {
  name: string | null | undefined;
  size?: keyof typeof sizes;
  tone?: keyof typeof tones;
  className?: string;
}) {
  return (
    <span
      aria-hidden="true"
      className={cn('flex shrink-0 items-center justify-center font-semibold', sizes[size], tones[tone], className)}
    >
      {initials(name)}
    </span>
  );
}
