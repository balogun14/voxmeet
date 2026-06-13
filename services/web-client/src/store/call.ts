import { create } from "zustand";

interface CallState {
  localStream: MediaStream | null;
  remoteStreams: Map<string, MediaStream>;
  isMuted: boolean;
  isVideoOff: boolean;
  isScreenSharing: boolean;
  peers: string[];
  connectionStatus: string;

  setLocalStream: (stream: MediaStream | null) => void;
  addRemoteStream: (userId: string, stream: MediaStream) => void;
  removeRemoteStream: (userId: string) => void;
  setMuted: (muted: boolean) => void;
  setVideoOff: (off: boolean) => void;
  setScreenSharing: (sharing: boolean) => void;
  setPeers: (peers: string[]) => void;
  addPeer: (userId: string) => void;
  removePeer: (userId: string) => void;
  setConnectionStatus: (status: string) => void;
  cleanup: () => void;
}

export const useCallStore = create<CallState>((set) => ({
  localStream: null,
  remoteStreams: new Map(),
  isMuted: false,
  isVideoOff: false,
  isScreenSharing: false,
  peers: [],
  connectionStatus: "disconnected",

  setLocalStream: (stream) => set({ localStream: stream }),
  addRemoteStream: (userId, stream) =>
    set((s) => {
      const next = new Map(s.remoteStreams);
      next.set(userId, stream);
      return { remoteStreams: next };
    }),
  removeRemoteStream: (userId) =>
    set((s) => {
      const next = new Map(s.remoteStreams);
      next.delete(userId);
      return { remoteStreams: next };
    }),
  setMuted: (muted) => set({ isMuted: muted }),
  setVideoOff: (off) => set({ isVideoOff: off }),
  setScreenSharing: (sharing) => set({ isScreenSharing: sharing }),
  setPeers: (peers) => set({ peers }),
  addPeer: (userId) =>
    set((s) => ({
      peers: s.peers.includes(userId) ? s.peers : [...s.peers, userId],
    })),
  removePeer: (userId) =>
    set((s) => ({ peers: s.peers.filter((id) => id !== userId) })),
  setConnectionStatus: (status) => set({ connectionStatus: status }),
  cleanup: () =>
    set({
      localStream: null,
      remoteStreams: new Map(),
      isMuted: false,
      isVideoOff: false,
      isScreenSharing: false,
      peers: [],
      connectionStatus: "disconnected",
    }),
}));
