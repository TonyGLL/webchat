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
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        type="text"
        value={name}
        onChange={e => setName(e.target.value)}
        placeholder="New room name"
      />
      <div className="flex items-center space-x-2">
        <Checkbox id="public-room" checked={isPublic} onCheckedChange={setIsPublic} />
        <label
          htmlFor="public-room"
          className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Public
        </label>
      </div>
      <Button type="submit" className="w-full">
        Create Room
      </Button>
    </form>
  );
}