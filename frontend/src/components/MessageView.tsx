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
      <div className="flex flex-col items-center justify-center h-full bg-gray-100 dark:bg-gray-800">
        <div className="text-center">
          <h2 className="text-2xl font-semibold text-gray-800 dark:text-gray-200">Welcome to Chat</h2>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Select a room to start messaging.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-gray-100 dark:bg-gray-800">
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
                        ? 'bg-blue-500 text-white'
                        : 'bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-200 shadow-md'
                    }`}
                  >
                    <p className="text-sm">{msg.content}</p>
                    <p className={`text-xs mt-1 text-right ${isCurrentUser ? 'text-blue-100' : 'text-gray-400'}`}>
                      {formatTimestamp(msg.created_at)}
                    </p>
                  </div>
                </div>
              );
            })}
          <div ref={messagesEndRef} />
        </div>
      </div>
      <div className="p-4 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700">
        <form onSubmit={handleSendMessage} className="flex items-center">
          <input
            type="text"
            value={content}
            onChange={e => setContent(e.target.value)}
            placeholder="Type a message..."
            className="w-full p-3 bg-gray-200 dark:bg-gray-800 border-transparent rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            className="ml-3 px-5 py-3 bg-blue-500 text-white rounded-full hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-blue-300"
            disabled={!content.trim()}
          >
            Send
          </button>
        </form>
      </div>
    </div>
  );
}