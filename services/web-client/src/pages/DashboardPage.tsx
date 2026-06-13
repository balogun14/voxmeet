import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { listRooms, createRoom } from "../api/client";

interface Room {
  id: string;
  name: string;
  is_public: boolean;
  created_at: string;
}

export default function DashboardPage() {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState("");
  const navigate = useNavigate();
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    listRooms(token).then(setRooms).catch(() => navigate("/login"));
  }, [token, navigate]);

  const handleCreate = async () => {
    if (!newName.trim() || !token) return;
    const room = await createRoom(token, newName, true);
    navigate(`/room/${room.id}`);
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">VoxMeet</h1>
        <button
          className="bg-blue-600 rounded px-4 py-2 text-sm"
          onClick={() => setShowCreate(!showCreate)}
        >
          New Room
        </button>
      </div>

      {showCreate && (
        <div className="flex gap-2 mb-4">
          <input
            className="bg-neutral-800 rounded px-3 py-2 flex-1"
            placeholder="Room name"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
          />
          <button
            className="bg-green-600 rounded px-4 py-2 text-sm"
            onClick={handleCreate}
          >
            Create
          </button>
        </div>
      )}

      <div className="flex flex-col gap-2">
        {rooms.map((room) => (
          <button
            key={room.id}
            className="bg-neutral-800 rounded p-4 text-left hover:bg-neutral-700 transition"
            onClick={() => navigate(`/room/${room.id}`)}
          >
            <span className="font-medium">{room.name}</span>
          </button>
        ))}
        {rooms.length === 0 && (
          <p className="text-neutral-500 text-sm">No rooms yet.</p>
        )}
      </div>
    </div>
  );
}
