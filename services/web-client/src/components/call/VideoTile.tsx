import { useRef, useEffect } from "react";

interface VideoTileProps {
  stream: MediaStream | null;
  muted?: boolean;
  label: string;
  mirrored?: boolean;
  className?: string;
}

export default function VideoTile({
  stream,
  muted = false,
  label,
  mirrored = false,
  className = "",
}: VideoTileProps) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (videoRef.current && stream) {
      videoRef.current.srcObject = stream;
    }
  }, [stream]);

  return (
    <div
      className={`relative bg-neutral-800 rounded-lg overflow-hidden ${className}`}
    >
      {stream ? (
        <video
          ref={videoRef}
          autoPlay
          playsInline
          muted={muted}
          className={`w-full h-full object-cover ${
            mirrored ? "-scale-x-100" : ""
          }`}
        />
      ) : (
        <div className="w-full h-full flex items-center justify-center text-neutral-500">
          No Video
        </div>
      )}
      <div className="absolute bottom-2 left-2 bg-black/60 rounded px-2 py-0.5 text-xs">
        {label}
        {muted && " (you)"}
      </div>
    </div>
  );
}
