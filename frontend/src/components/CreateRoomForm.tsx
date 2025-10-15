'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';

export default function CreateRoomForm({
  createRoom,
}: {
  createRoom: (data: { name: string; is_public: boolean }) => void;
}) {
  const [name, setName] = useState('');
  const [isPublic, setIsPublic] = useState(true);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (name.trim()) {
      createRoom({ name, is_public: isPublic });
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
        className="w-full p-3 border border-gray-300 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
      />
      <div className="flex items-center space-x-2">
        <Checkbox id="public-room" checked={isPublic} onCheckedChange={() => setIsPublic(!isPublic)} />
        <label
          htmlFor="public-room"
          className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 text-gray-800 dark:text-gray-200"
        >
          Public Room
        </label>
      </div>
      <Button
        type="submit"
        className="w-full bg-blue-500 text-white rounded-full hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-blue-300"
        disabled={!name.trim()}
      >
        Create
      </Button>
    </form>
  );
}