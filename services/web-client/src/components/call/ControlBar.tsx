import { useCallStore } from "../../store/call";

interface ControlBarProps {
  onLeave: () => void;
  onScreenShare?: () => void;
}

export default function ControlBar({ onLeave, onScreenShare }: ControlBarProps) {
  const { isMuted, isVideoOff, isScreenSharing, setMuted, setVideoOff, setScreenSharing } = useCallStore();

  return (
    <div className="bg-neutral-900 px-4 py-3 flex items-center justify-center gap-4">
      <button
        onClick={() => setMuted(!isMuted)}
        className={`rounded-full p-3 transition ${
          isMuted ? "bg-red-600" : "bg-neutral-700 hover:bg-neutral-600"
        }`}
        title={isMuted ? "Unmute" : "Mute"}
      >
        {isMuted ? (
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" />
          </svg>
        ) : (
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
          </svg>
        )}
      </button>

      <button
        onClick={() => setVideoOff(!isVideoOff)}
        className={`rounded-full p-3 transition ${
          isVideoOff ? "bg-red-600" : "bg-neutral-700 hover:bg-neutral-600"
        }`}
        title={isVideoOff ? "Turn on camera" : "Turn off camera"}
      >
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
      </button>

      <button
        onClick={onLeave}
        className="bg-red-600 rounded-full px-6 py-2 text-sm font-medium hover:bg-red-700 transition"
      >
        Leave
      </button>

      {onScreenShare && (
        <button
          onClick={() => {
            onScreenShare();
            setScreenSharing(!isScreenSharing);
          }}
          className={`rounded-full p-3 transition ${
            isScreenSharing ? "bg-blue-600" : "bg-neutral-700 hover:bg-neutral-600"
          }`}
          title={isScreenSharing ? "Stop sharing" : "Share screen"}
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
        </button>
      )}
    </div>
  );
}
