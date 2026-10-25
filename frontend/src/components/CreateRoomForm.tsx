'use client';

import { useState } from 'react';
import { Plus, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { useT } from '@/lib/i18n/useT';
import { logError } from '@/lib/logError';

export default function CreateRoomForm({
  createRoom,
}: {
  createRoom: (data: { name: string; is_private: boolean; topic: string | null }) => Promise<void> | void;
}) {
  const { t } = useT();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [topic, setTopic] = useState('');
  const [isPublic, setIsPublic] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [failed, setFailed] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || submitting) return;
    setSubmitting(true);
    setFailed(false);
    try {
      await createRoom({ name: name.trim(), is_private: !isPublic, topic: topic.trim() || null });
      setName('');
      setTopic('');
      setOpen(false);
    } catch (err) {
      logError('Failed to create room', err);
      setFailed(true);
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
        aria-label={open ? t('rooms.closeForm') : t('rooms.new')}
        aria-expanded={open}
        title={t('rooms.new')}
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
            onChange={(e) => {
              setName(e.target.value);
              setFailed(false);
            }}
            placeholder={t('rooms.name')}
            aria-label={t('rooms.name')}
            required
          />
          <Input
            value={topic}
            onChange={(e) => setTopic(e.target.value)}
            maxLength={255}
            placeholder={t('rooms.topic')}
            aria-label={t('rooms.topic')}
          />
          {failed && (
            <p role="alert" className="text-xs text-destructive">
              {t('rooms.createError')}
            </p>
          )}
          <div className="flex items-center justify-between">
            <label htmlFor="public-room" className="flex items-center gap-2 text-sm">
              <Checkbox id="public-room" checked={isPublic} onCheckedChange={() => setIsPublic(!isPublic)} />
              {t('rooms.public')}
            </label>
            <Button type="submit" size="sm" disabled={!name.trim() || submitting}>
              {t('rooms.create')}
            </Button>
          </div>
        </form>
      )}
    </>
  );
}
