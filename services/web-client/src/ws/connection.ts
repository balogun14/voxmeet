type MessageHandler = (msg: ServerMessage) => void;
type StatusHandler = (status: ConnectionStatus) => void;

export type ConnectionStatus = "disconnected" | "connecting" | "connected";

export interface ServerMessage {
  type: string;
  room_id?: string;
  user_id?: string;
  data?: Record<string, unknown>;
}

export class WSConnection {
  private ws: WebSocket | null = null;
  private url: string;
  private onMessage: MessageHandler;
  private onStatusChange: StatusHandler;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private _status: ConnectionStatus = "disconnected";

  constructor(
    token: string,
    onMessage: MessageHandler,
    onStatusChange: StatusHandler
  ) {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    this.url = `${protocol}//${window.location.host}/api/v1/ws?token=${token}`;
    this.onMessage = onMessage;
    this.onStatusChange = onStatusChange;
  }

  get status() {
    return this._status;
  }

  connect() {
    this.setStatus("connecting");
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => this.setStatus("connected");
    this.ws.onclose = () => {
      this.setStatus("disconnected");
      this.scheduleReconnect();
    };
    this.ws.onmessage = (event) => {
      try {
        const msg: ServerMessage = JSON.parse(event.data);
        this.onMessage(msg);
      } catch {
        // ignore malformed messages
      }
    };
  }

  send(msg: Record<string, unknown>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  disconnect() {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.ws?.close();
    this.ws = null;
    this.setStatus("disconnected");
  }

  private setStatus(status: ConnectionStatus) {
    this._status = status;
    this.onStatusChange(status);
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, 3000);
  }
}
