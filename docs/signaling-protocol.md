# Signaling Protocol

All real-time communication uses a single WebSocket connection per client to `api-gateway`.

## Connection

```
ws://host:8080/api/v1/ws?token=<JWT>
```

The JWT is obtained from `POST /api/v1/auth/login`. The server responds with `{"type":"authenticated","user_id":"..."}` on successful connection.

## Message Format

All messages are JSON with a `type` field.

### Client → Server

```json
{"type":"join_room","room_id":"uuid","data":{...}}
{"type":"sdp_answer","room_id":"uuid","data":{"sdp":"...","type":"answer"}}
{"type":"ice_candidate","room_id":"uuid","data":{"candidate":"...","sdp_mid":"0","sdp_mline_index":0}}
{"type":"publish","room_id":"uuid","data":{"track_id":"...","kind":"audio","source":"mic"}}
{"type":"unpublish","room_id":"uuid","data":{"track_id":"..."}}
{"type":"leave_room","room_id":"uuid"}
{"type":"chat.send","room_id":"uuid","data":{"content":"Hello!"}}
{"type":"chat.typing","room_id":"uuid","data":{"is_typing":true}}
{"type":"ping"}
```

### Server → Client

```json
{"type":"authenticated","user_id":"..."}
{"type":"room_state","room_id":"uuid","data":{"peers":[...],"tracks":[...]}}
{"type":"peer_joined","room_id":"uuid","data":{"user_id":"...","tracks":[...]}}
{"type":"peer_left","room_id":"uuid","data":{"user_id":"..."}}
{"type":"sdp_offer","room_id":"uuid","data":{"sdp":"...","type":"offer"}}
{"type":"ice_candidate","room_id":"uuid","data":{"candidate":"...","sdp_mid":"0","sdp_mline_index":0}}
{"type":"new_track","room_id":"uuid","data":{"publisher_id":"...","track_id":"...","kind":"video","source":"camera"}}
{"type":"remove_track","room_id":"uuid","data":{"track_id":"..."}}
{"type":"chat.message","room_id":"uuid","data":{"user_id":"...","username":"...","content":"...","timestamp":"..."}}
{"type":"chat.typing","room_id":"uuid","data":{"user_id":"...","is_typing":true}}
{"type":"error","code":"...","message":"..."}
{"type":"pong"}
```

## Signaling Flow (Group Call)

```
Client                    api-gateway              RabbitMQ                  SFU
  │                           │                       │                       │
  │── join_room ──────────────>──────────────────────>──────────────────────>│
  │                           │                       │                       │── Create PeerConnection
  │                           │                       │                       │── Generate SDP offer
  │<── sdp_offer ────────────<──────────────────────<────────────────────────│
  │                                                                           │
  │── sdp_answer ────────────>──────────────────────>──────────────────────>│── Set remote SDP
  │                                                                           │
  │── ice_candidate ─────────>──────────────────────>──────────────────────>│── Add ICE candidate
  │<── ice_candidate ────────<──────────────────────<────────────────────────│
  │                                                                           │
  │══════════════ Media flows directly Client ↔ SFU (RTP/RTCP) ═══════════════>│
  │                                                                           │
  │── publish ───────────────>──────────────────────>──────────────────────>│── Register track
  │                                                                           │── Route to subscribers
  │<── new_track (from other peers) ──────────────────────────────────────────│
  │                                                                           │
  │── leave_room ────────────>──────────────────────>──────────────────────>│── Cleanup
```

## Error Codes

| Code | Description |
|---|---|
| `invalid_message` | Malformed JSON |
| `unknown_type` | Unrecognized message type |
| `unauthorized` | Missing or invalid token |
| `room_not_found` | Room doesn't exist |
| `room_full` | Room has reached max participants |
