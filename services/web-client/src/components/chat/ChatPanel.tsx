import { useState, useRef, useEffect } from "react";
import { useChatStore } from "../../store/chat";

export default function ChatPanel() {
  const [input, setInput] = useState("");
  const messages = useChatStore((s) => s.messages);
  const addMessage = useChatStore((s) => s.addMessage);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;
    addMessage({
      id: crypto.randomUUID(),
      userId: "me",
      username: "You",
      content: input,
      timestamp: new Date().toISOString(),
    });
    setInput("");
  };

  return (
    <div className="w-80 bg-neutral-900 flex flex-col border-l border-neutral-800">
      <div className="px-4 py-3 border-b border-neutral-800 font-medium text-sm">
        Chat
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.length === 0 && (
          <p className="text-neutral-500 text-xs text-center">
            No messages yet
          </p>
        )}
        {messages.map((msg) => (
          <div key={msg.id} className="text-sm">
            <span className="font-medium text-neutral-300">
              {msg.username}
            </span>
            <span className="text-neutral-500 text-xs ml-2">
              {new Date(msg.timestamp).toLocaleTimeString()}
            </span>
            <p className="text-neutral-100 mt-0.5">{msg.content}</p>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <form onSubmit={handleSend} className="p-3 border-t border-neutral-800">
        <input
          className="w-full bg-neutral-800 rounded px-3 py-2 text-sm outline-none focus:ring-1 focus:ring-blue-500"
          placeholder="Send a message..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
        />
      </form>
    </div>
  );
}
