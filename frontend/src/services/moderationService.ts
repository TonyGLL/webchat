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

export interface Report {
  id: number;
  message_id: string;
  room_id: string;
  message_content: string;
  message_author_id: number | null;
  reported_by_id: number;
  reason: string;
  status: 'pending' | 'resolved' | 'ignored';
  created_at: string;
}

export interface Ban {
  user_id: number;
  reason: string;
  expires_at: string | null;
  created_at: string;
}

export const moderationService = {
  report: async (messageId: string, reason: string): Promise<void> => {
    await api.post(`/moderation/messages/${messageId}/report`, { reason });
  },

  getReports: async (roomId: string): Promise<Report[]> => {
    const response = await api.get(`/moderation/rooms/${roomId}/reports`);
    return response.data ?? [];
  },

  resolveReport: async (reportId: number, action: 'resolve' | 'ignore', deleteMessage = false): Promise<void> => {
    await api.post(`/moderation/reports/${reportId}/resolve`, { action, delete_message: deleteMessage });
  },

  ban: async (roomId: string, userId: number, reason: string, duration?: string): Promise<void> => {
    await api.post(`/moderation/rooms/${roomId}/bans/${userId}`, { reason, duration });
  },

  unban: async (roomId: string, userId: number): Promise<void> => {
    await api.delete(`/moderation/rooms/${roomId}/bans/${userId}`);
  },

  getBans: async (roomId: string): Promise<Ban[]> => {
    const response = await api.get(`/moderation/rooms/${roomId}/bans`);
    return response.data ?? [];
  },

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
