import { useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getRoom } from "../api/client";
import VideoGrid from "../components/call/VideoGrid";
import ControlBar from "../components/call/ControlBar";
import ChatPanel from "../components/chat/ChatPanel";
import { SignalingManager } from "../ws/signaling";
import { useCallStore } from "../store/call";
import { getLocalStream, stopStream, getScreenStream } from "../webrtc/media";

interface RoomInfo {
  id: string;
  name: string;
}

export default function RoomPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [room, setRoom] = useState<RoomInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [screenShareStream, setScreenShareStream] = useState<MediaStream | null>(null);
  const signalingRef = useRef<SignalingManager | null>(null);
  const {
    setLocalStream,
    addRemoteStream,
    removeRemoteStream,
    addPeer,
    setPeers,
    removePeer,
    cleanup,
    localStream,
  } = useCallStore();

  const token = localStorage.getItem("token");
  const userId = localStorage.getItem("user_id") || "anon";

  useEffect(() => {
    if (!token || !id) {
      navigate("/login");
      return;
    }

    getRoom(token, id)
      .then(setRoom)
      .catch(() => navigate("/"))
      .finally(() => setLoading(false));

    // Get local media
    let stream: MediaStream | null = null;
    getLocalStream(true, true)
      .then((s) => {
        stream = s;
        setLocalStream(s);
      })
      .catch(() => {
        // User denied permissions — proceed without media
      });

    // Set up signaling
    const sig = new SignalingManager(token, id, userId, {
      onRemoteStream: (s, uid) => addRemoteStream(uid, s),
      onPeerJoined: (uid) => addPeer(uid),
      onPeerLeft: (uid) => {
        removePeer(uid);
        removeRemoteStream(uid);
      },
      onNewTrack: () => {},
      onRoomState: (peers) => setPeers(peers),
      onError: () => {},
    });
    signalingRef.current = sig;

    // Connect signaling after a short delay to let local stream init
    setTimeout(() => sig.connect(stream), 500);

    return () => {
      sig.disconnect();
      if (stream) stopStream(stream);
      cleanup();
    };
  }, [id, token, userId, navigate]);

  const handleLeave = () => {
    signalingRef.current?.disconnect();
    if (localStream) stopStream(localStream);
    cleanup();
    navigate("/");
  };

  const handleScreenShare = async () => {
    if (screenShareStream) {
      stopStream(screenShareStream);
      setScreenShareStream(null);
      return;
    }
    try {
      const stream = await getScreenStream();
      setScreenShareStream(stream);
      if (signalingRef.current) {
        // Replace local tracks with screen share
        // In real SFU, this would be a new publish action
      }
    } catch {
      // User cancelled or not supported
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        Loading...
      </div>
    );
  }

  if (!room) return null;

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-neutral-900 px-6 py-3 flex items-center justify-between border-b border-neutral-800">
        <h1 className="text-lg font-bold">{room.name}</h1>
        <span className="text-xs text-neutral-400">{userId.slice(0, 8)}</span>
      </header>

      <div className="flex-1 flex">
        <div className="flex-1 flex flex-col">
          <VideoGrid />
          <ControlBar onLeave={handleLeave} onScreenShare={handleScreenShare} />
        </div>
        <ChatPanel />
      </div>
    </div>
  );
}
