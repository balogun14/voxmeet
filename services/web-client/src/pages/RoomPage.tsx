import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getRoom } from "../api/client";
import { WSConnection, ServerMessage } from "../ws/connection";

interface RoomInfo {
  id: string;
  name: string;
}

export default function RoomPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [room, setRoom] = useState<RoomInfo | null>(null);
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token || !id) {
      navigate("/login");
      return;
    }

    getRoom(token, id)
      .then(setRoom)
      .catch(() => navigate("/"));

    const onMessage = (msg: ServerMessage) => {
      console.log("WS message:", msg);
    };

    const ws = new WSConnection(token, onMessage, console.log);
    ws.connect();

    // Join room via WebSocket
    ws.send({ type: "join_room", room_id: id });

    return () => ws.disconnect();
  }, [id, token, navigate]);

  if (!room) {
    return <div className="p-6">Loading...</div>;
  }

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-neutral-900 px-6 py-3 flex items-center justify-between">
        <h1 className="text-lg font-bold">{room.name}</h1>
        <button
          className="text-sm text-neutral-400"
          onClick={() => navigate("/")}
        >
          Leave
        </button>
      </header>
      <main className="flex-1 flex items-center justify-center text-neutral-600">
        <p>Join a call to see video here.</p>
      </main>
    </div>
  );
}
