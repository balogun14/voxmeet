export interface PeerConfig {
  iceServers: RTCIceServer[];
  onTrack: (stream: MediaStream, userId: string) => void;
  onICECandidate: (candidate: RTCIceCandidateInit) => void;
  onStateChange: (state: RTCPeerConnectionState) => void;
}

export class PeerConnectionManager {
  private pc: RTCPeerConnection;
  private userId: string;
  private onTrack: (stream: MediaStream, userId: string) => void;
  private onICECandidate: (candidate: RTCIceCandidateInit) => void;

  constructor(userId: string, config: PeerConfig) {
    this.userId = userId;
    this.onTrack = config.onTrack;
    this.onICECandidate = config.onICECandidate;

    this.pc = new RTCPeerConnection({
      iceServers: config.iceServers,
    });

    this.pc.onicecandidate = (evt) => {
      if (evt.candidate) {
        this.onICECandidate(evt.candidate.toJSON());
      }
    };

    this.pc.ontrack = (evt) => {
      this.onTrack(evt.streams[0], this.userId);
    };

    this.pc.onconnectionstatechange = () => {
      config.onStateChange(this.pc.connectionState);
    };
  }

  async createOffer(): Promise<RTCSessionDescriptionInit> {
    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);
    return offer;
  }

  async setAnswer(sdp: string): Promise<void> {
    await this.pc.setRemoteDescription(
      new RTCSessionDescription({ type: "answer", sdp })
    );
  }

  async setOffer(sdp: string): Promise<RTCSessionDescriptionInit> {
    await this.pc.setRemoteDescription(
      new RTCSessionDescription({ type: "offer", sdp })
    );
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);
    return answer;
  }

  async addICECandidate(candidate: RTCIceCandidateInit): Promise<void> {
    await this.pc.addIceCandidate(new RTCIceCandidate(candidate));
  }

  async addLocalStream(stream: MediaStream): Promise<void> {
    for (const track of stream.getTracks()) {
      this.pc.addTrack(track, stream);
    }
  }

  close(): void {
    this.pc.close();
  }

  getConnectionState(): RTCPeerConnectionState {
    return this.pc.connectionState;
  }
}
