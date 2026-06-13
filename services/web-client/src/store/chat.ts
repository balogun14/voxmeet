import { create } from "zustand";

interface ChatMessage {
  id: string;
  userId: string;
  username: string;
  content: string;
  timestamp: string;
}

interface ChatState {
  messages: ChatMessage[];
  addMessage: (msg: ChatMessage) => void;
  clearMessages: () => void;
}

export const useChatStore = create<ChatState>((set) => ({
  messages: [],
  addMessage: (msg) =>
    set((s) => ({ messages: [...s.messages, msg] })),
  clearMessages: () => set({ messages: [] }),
}));
