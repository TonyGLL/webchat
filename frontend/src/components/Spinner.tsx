import { Loader2 } from 'lucide-react';

export default function Spinner({ className = 'h-5 w-5' }: { className?: string }) {
  return (
    <div className="flex items-center justify-center" role="status" aria-label="Loading">
      <Loader2 className={`animate-spin text-muted-foreground ${className}`} />
    </div>
  );
}
