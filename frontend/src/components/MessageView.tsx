import { useState } from 'react';
import { useMessages } from '@/hooks/useMessages';
import Spinner from './Spinner';

export default function MessageView({ roomId, socket }: { roomId: string | null, socket: any }) {
  const { messages, loading, createMessage, addReaction } = useMessages(roomId, socket);
  const [content, setContent] = useState('');

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (roomId && content) {
      createMessage({ room_id: roomId, content });
      setContent('');
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 p-4 overflow-y-auto">
        {loading && <Spinner />}
        {messages.map((msg) => (
          <div key={msg.id} className="mb-4">
            <span className="font-bold">{msg.author.username}: </span>
            <span>{msg.content}</span>
            <div className="flex">
              {msg.reactions &&
                Object.entries(msg.reactions).map(([emoji, count]) => (
                  <div key={emoji} className="mr-2">
                    {emoji} {count as number}
                  </div>
                ))}
              <button onClick={() => addReaction(msg.id, '👍')} className="ml-2">
                👍
              </button>
            </div>
          </div>
        ))}
      </div>
      <form onSubmit={handleSendMessage} className="p-4 border-t border-gray-200">
        <input
          type="text"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Type a message"
          className="w-full p-2 border border-gray-300 rounded"
        />
      </form>
    </div>
  );
}