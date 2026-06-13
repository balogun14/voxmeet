import VideoTile from "./VideoTile";
import { useCallStore } from "../../store/call";

export default function VideoGrid() {
  const { localStream, remoteStreams, peers } = useCallStore();

  const allPeers = ["local", ...peers];
  const cols = allPeers.length <= 2 ? 2 : 3;

  return (
    <div
      className="grid gap-2 p-2 flex-1"
      style={{
        gridTemplateColumns: `repeat(${Math.min(cols, 3)}, 1fr)`,
      }}
    >
      <VideoTile
        stream={localStream}
        muted
        label="You"
        mirrored
      />
      {peers.map((userId) => (
        <VideoTile
          key={userId}
          stream={remoteStreams.get(userId) || null}
          label={userId.slice(0, 8)}
        />
      ))}
      {peers.length === 0 && (
        <div className="col-span-full flex items-center justify-center text-neutral-500 text-sm">
          Waiting for others to join...
        </div>
      )}
    </div>
  );
}
