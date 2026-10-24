import { api } from '@/lib/api';
import { RoomRole } from '@/lib/permissions';

export interface AuditEntry {
  id: number;
  actor_id: number | null;
  action: string;
  target_user_id?: number;
  details: Record<string, unknown>;
  created_at: string;
}

export const moderationService = {
  changeRole: async (roomId: string, userId: number, role: RoomRole): Promise<void> => {
    await api.patch(`/moderation/rooms/${roomId}/members/${userId}/role`, { role });
  },

  kick: async (roomId: string, userId: number, reason?: string): Promise<void> => {
    await api.delete(`/moderation/rooms/${roomId}/members/${userId}`, { data: { reason } });
  },

  getAudit: async (roomId: string): Promise<AuditEntry[]> => {
    const response = await api.get(`/moderation/rooms/${roomId}/audit`);
    return response.data ?? [];
  },
};
