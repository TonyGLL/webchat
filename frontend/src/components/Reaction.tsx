import data from '@emoji-mart/data'
import Picker from '@emoji-mart/react'
import { useState } from 'react';

interface ReactionProps {
  messageId: string;
  reactions: { user_id: number; emoji: string }[];
  onAddReaction: (messageId: string, emoji: string) => void;
  onRemoveReaction: (messageId: string, emoji: string) => void;
  currentUserId: number;
}

export default function Reaction({ messageId, reactions, onAddReaction, onRemoveReaction, currentUserId }: ReactionProps) {
  const [showPicker, setShowPicker] = useState(false);

  const handleEmojiSelect = (emoji: any) => {
    const existingReaction = reactions.find((r) => r.emoji === emoji.native && r.user_id === currentUserId);
    if (existingReaction) {
      onRemoveReaction(messageId, emoji.native);
    } else {
      onAddReaction(messageId, emoji.native);
    }
    setShowPicker(false);
  };

  const groupReactions = (reactions: { user_id: number; emoji: string }[]) => {
    return reactions.reduce((acc, reaction) => {
      const existing = acc.find((item) => item.emoji === reaction.emoji);
      if (existing) {
        existing.count++;
        existing.users.push(reaction.user_id);
      } else {
        acc.push({ emoji: reaction.emoji, count: 1, users: [reaction.user_id] });
      }
      return acc;
    }, [] as { emoji: string; count: number; users: number[] }[]);
  };

  const groupedReactions = groupReactions(reactions);

  return (
    <div className="relative">
      {showPicker && (
        <div>
          <Picker data={data} onEmojiSelect={handleEmojiSelect} />
        </div>
      )}
      <div onClick={() => setShowPicker(!showPicker)} className="flex items-center gap-1 justify-end">
        {groupedReactions.map((reaction, index) => (
          <div
            key={index}
            className={`px-2 py-1 rounded-full cursor-pointer ${reaction.users.includes(currentUserId) ? 'bg-blue-500' : 'bg-gray-700'}`}>
            {reaction.emoji} {reaction.count > 1 && <span>{reaction.count}</span>}
          </div>
        ))}
      </div>
    </div>
  );
}
