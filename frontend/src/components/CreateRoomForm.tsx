'use client';

import { useState } from 'react';
import { Plus, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';

export default function CreateRoomForm({
  createRoom,
}: {
  createRoom: (data: { name: string; is_private: boolean; topic: string }) => Promise<void> | void;
}) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [isPublic, setIsPublic] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || submitting) return;
    setSubmitting(true);
    try {
      await createRoom({ name: name.trim(), is_private: !isPublic, topic: 'Games' });
      setName('');
      setOpen(false);
    } catch (err) {
      console.error('Failed to create room', err);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        className="h-8 w-8"
        onClick={() => setOpen((v) => !v)}
        aria-label={open ? 'Close new room form' : 'New room'}
        aria-expanded={open}
        title="New room"
      >
        {open ? <X className="h-4 w-4" /> : <Plus className="h-4 w-4" />}
      </Button>
      {open && (
        <form
          onSubmit={handleSubmit}
          className="absolute left-0 right-0 top-full z-10 space-y-3 border-b bg-background p-4 shadow-sm"
        >
          <Input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Room name"
            aria-label="Room name"
          />
          <div className="flex items-center justify-between">
            <label htmlFor="public-room" className="flex items-center gap-2 text-sm">
              <Checkbox id="public-room" checked={isPublic} onCheckedChange={() => setIsPublic(!isPublic)} />
              Public
            </label>
            <Button type="submit" size="sm" disabled={!name.trim() || submitting}>
              Create
            </Button>
          </div>
        </form>
      )}
    </>
  );
}
