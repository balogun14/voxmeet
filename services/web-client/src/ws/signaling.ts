import { WSConnection, ServerMessage } from "../ws/connection";
import {
  PeerConnectionManager,
  PeerConfig,
} from "../webrtc/peer";

type SignalingEvents = {
  onRemoteStream: (stream: MediaStream, userId: string) => void;
  onPeerJoined: (userId: string) => void;
  onPeerLeft: (userId: string) => void;
  onNewTrack: (publisherId: string, kind: string) => void;
  onRoomState: (peers: string[]) => void;
  onError: (code: string, msg: string) => void;
};

export class SignalingManager {
  private ws: WSConnection;
  private peers: Map<string, PeerConnectionManager> = new Map();
  private localStream: MediaStream | null = null;
  private roomId: string;
  private events: SignalingEvents;
  private iceServers: RTCIceServer[];
  private remoteStreams: Map<string, MediaStream> = new Map();

  constructor(
    token: string,
    roomId: string,
    _userId: string,
    events: SignalingEvents,
    iceServers?: RTCIceServer[]
  ) {
    this.roomId = roomId;
    this.events = events;
    this.iceServers = iceServers || [
      { urls: "stun:stun.l.google.com:19302" },
    ];

    this.ws = new WSConnection(
      token,
      this.handleMessage.bind(this),
      () => {}
    );
  }

  async connect(stream?: MediaStream | null) {
    this.localStream = stream || null;
    this.ws.connect();
  }

  disconnect() {
    this.leaveRoom();
    this.ws.disconnect();
    this.peers.forEach((p) => p.close());
    this.peers.clear();
    this.remoteStreams.forEach((s) =>
      s.getTracks().forEach((t) => t.stop())
    );
    this.remoteStreams.clear();
  }

  private handleMessage(msg: ServerMessage) {
    const data = msg.data || {};

    switch (msg.type) {
      case "authenticated":
        // Join room after auth
        this.ws.send({ type: "join_room", room_id: this.roomId });
        break;

      case "sdp_offer":
        this.handleOffer(msg.user_id!, (data as any).sdp);
        break;

      case "sdp_answer":
        this.handleAnswer(msg.user_id!, (data as any).sdp);
        break;

      case "ice_candidate":
        this.handleICE(msg.user_id!, data as any);
        break;

      case "room_state":
        this.handleRoomState(data as any);
        break;

      case "peer_joined":
        this.events.onPeerJoined(msg.user_id!);
        if (this.localStream) {
          this.createPeerConnection(msg.user_id!);
        }
        break;

      case "peer_left":
        this.events.onPeerLeft(msg.user_id!);
        this.removePeer(msg.user_id!);
        break;

      case "new_track":
        this.events.onNewTrack(
          (data as any).publisher_id,
          (data as any).kind
        );
        break;

      case "error":
        this.events.onError((data as any).code, (data as any).message);
        break;
    }
  }

  private async createPeerConnection(remoteUserId: string) {
    if (this.peers.has(remoteUserId)) return;

    const config: PeerConfig = {
      iceServers: this.iceServers,
      onTrack: (stream, uid) => {
        this.remoteStreams.set(uid, stream);
        this.events.onRemoteStream(stream, uid);
      },
      onICECandidate: (candidate) => {
        this.ws.send({
          type: "ice_candidate",
          room_id: this.roomId,
          data: candidate,
        });
      },
      onStateChange: (state) => {
        if (state === "disconnected" || state === "failed") {
          this.removePeer(remoteUserId);
        }
      },
    };

    const peer = new PeerConnectionManager(remoteUserId, config);
    this.peers.set(remoteUserId, peer);

    if (this.localStream) {
      await peer.addLocalStream(this.localStream);
    }

    // SFU creates the offer, so we just answer
    // But we need to create an offer so the SFU can answer
    // Actually: SFU sends us an offer, we answer.
    // The createPeerConnection is called when a new peer joins
    // so we can start sending our media to them via the SFU
  }

  private async handleOffer(senderId: string, sdp: string) {
    if (!this.peers.has(senderId)) {
      const config: PeerConfig = {
        iceServers: this.iceServers,
        onTrack: (stream, uid) => {
          this.remoteStreams.set(uid, stream);
          this.events.onRemoteStream(stream, uid);
        },
        onICECandidate: (candidate) => {
          this.ws.send({
            type: "ice_candidate",
            room_id: this.roomId,
            data: candidate,
          });
        },
        onStateChange: () => {},
      };
      const peer = new PeerConnectionManager(senderId, config);
      this.peers.set(senderId, peer);

      if (this.localStream) {
        await peer.addLocalStream(this.localStream);
      }
    }

    const peer = this.peers.get(senderId)!;
    const answer = await peer.setOffer(sdp);
    this.ws.send({
      type: "sdp_answer",
      room_id: this.roomId,
      data: { sdp: answer.sdp, type: "answer" },
    });
  }

  private async handleAnswer(senderId: string, sdp: string) {
    const peer = this.peers.get(senderId);
    if (peer) {
      await peer.setAnswer(sdp);
    }
  }

  private async handleICE(
    senderId: string,
    data: { candidate: string; sdp_mid?: string; sdp_mline_index?: number }
  ) {
    const peer = this.peers.get(senderId);
    if (peer && data.candidate) {
      await peer.addICECandidate({
        candidate: data.candidate,
        sdpMid: data.sdp_mid || null,
        sdpMLineIndex: data.sdp_mline_index ?? null,
      });
    }
  }

  private handleRoomState(data: { peers: Array<{ user_id: string }> }) {
    const peerIds = data.peers.map((p) => p.user_id);
    this.events.onRoomState(peerIds);
  }

  private removePeer(userId: string) {
    const peer = this.peers.get(userId);
    if (peer) {
      peer.close();
      this.peers.delete(userId);
    }
    this.remoteStreams.delete(userId);
  }

  private leaveRoom() {
    this.ws.send({ type: "leave_room", room_id: this.roomId });
  }
}
