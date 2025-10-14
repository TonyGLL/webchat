import { useState } from 'react';

export default function CreateRoomForm({ createRoom }: { createRoom: (data: { name: string; is_public: boolean }) => void }) {
  const [name, setName] = useState('');
  const [isPublic, setIsPublic] = useState(true);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createRoom({ name, is_public: isPublic });
    setName('');
  };

  return (
    <form onSubmit={handleSubmit} className="p-4 border-t border-gray-200">
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="New room name"
        className="w-full p-2 border border-gray-300 rounded"
      />
      <div className="flex items-center my-2">
        <input
          type="checkbox"
          checked={isPublic}
          onChange={(e) => setIsPublic(e.target.checked)}
          className="mr-2"
        />
        <label>Public</label>
      </div>
      <button type="submit" className="w-full p-2 bg-blue-500 text-white rounded">
        Create Room
      </button>
    </form>
  );
}