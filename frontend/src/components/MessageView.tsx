import { useState, useEffect, useRef } from 'react';
import { useMessages } from '@/hooks/useMessages';
import Spinner from './Spinner';
import { useUser } from '@/hooks/useUser';

export default function MessageView({ roomId }: { roomId: string | null }) {
  const { messages, loading, createMessage } = useMessages(roomId);
  const { profile } = useUser();
  const [content, setContent] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (roomId && content.trim()) {
      createMessage({ room_id: roomId, content });
      setContent('');
    }
  };

  const formatTimestamp = (date: Date) => {
    return new Date(date).toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: 'numeric',
      hour12: true,
    });
  };

  if (!roomId) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-transparent">
        <div className="text-center">
          <h2 className="text-2xl font-semibold text-white">Welcome to Chat</h2>
          <p className="mt-2 text-gray-400">Select a room to start messaging.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-transparent">
      <div className="flex-1 p-6 overflow-y-auto">
        <div className="flex flex-col space-y-4">
          {loading && (
            <div className="flex justify-center items-center h-full">
              <Spinner />
            </div>
          )}
          {!loading &&
            messages.map(msg => {
              const isCurrentUser = msg.author_id === profile?.id;
              return (
                <div key={msg.id} className={`flex items-end ${isCurrentUser ? 'justify-end' : 'justify-start'}`}>
                  <div
                    className={`max-w-xs lg:max-w-md px-4 py-2 rounded-2xl ${
                      isCurrentUser
                        ? 'bg-whatsapp-gossip text-gray-800'
                        : 'bg-white text-gray-800 shadow-md'
                    }`}
                  >
                    <p className="text-sm">{msg.content}</p>
                    <p className={`text-xs mt-1 text-right ${isCurrentUser ? 'text-gray-500' : 'text-gray-400'}`}>
                      {formatTimestamp(msg.created_at)}
                    </p>
                  </div>
                </div>
              );
            })}
          <div ref={messagesEndRef} />
        </div>
      </div>
      <div className="p-4 bg-whatsapp-surfie-green border-t border-gray-700">
        <form onSubmit={handleSendMessage} className="flex items-center">
          <input
            type="text"
            value={content}
            onChange={e => setContent(e.target.value)}
            placeholder="Type a message..."
            className="w-full p-3 bg-whatsapp-deep-sea-green border-transparent rounded-full focus:outline-none focus:ring-2 focus:ring-whatsapp-mountain-meadow text-white placeholder-gray-400"
          />
          <button
            type="submit"
            className="ml-3 px-5 py-3 bg-whatsapp-mountain-meadow text-white rounded-full hover:bg-opacity-80 focus:outline-none focus:ring-2 focus:ring-whatsapp-mountain-meadow disabled:bg-opacity-50"
            disabled={!content.trim()}
          >
            Send
          </button>
        </form>
      </div>
    </div>
  );
}