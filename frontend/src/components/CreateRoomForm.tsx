'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';

export default function CreateRoomForm({
  createRoom,
}: {
  createRoom: (data: { name: string; is_public: boolean, topic: string }) => void;
}) {
  const [name, setName] = useState('');
  const [isPublic, setIsPublic] = useState(true);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (name.trim()) {
      createRoom({ name, is_public: isPublic, topic: 'Games' });
      setName('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 p-4">
      <Input
        type="text"
        value={name}
        onChange={e => setName(e.target.value)}
        placeholder="Create a new room"
        className="w-full p-3 border-gray-600 rounded-full focus:outline-none focus:ring-2 focus:ring-whatsapp-mountain-meadow bg-whatsapp-deep-sea-green text-white placeholder-gray-400"
      />
      <div className="flex items-center space-x-2">
        <Checkbox id="public-room" checked={isPublic} onCheckedChange={() => setIsPublic(!isPublic)} />
        <label
          htmlFor="public-room"
          className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 text-white"
        >
          Public Room
        </label>
      </div>
      <Button
        type="submit"
        className="w-full bg-whatsapp-mountain-meadow text-white rounded-full hover:bg-opacity-80 focus:outline-none focus:ring-2 focus:ring-whatsapp-mountain-meadow disabled:bg-opacity-50"
        disabled={!name.trim()}
      >
        Create
      </Button>
    </form>
  );
}