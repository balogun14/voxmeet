# VoxMeet — UAT Test Guideline

## Prerequisites

```bash
# Clone, configure, start everything
git clone https://github.com/your-org/voxmeet.git
cd voxmeet
cp .env.example .env
# Edit .env — set JWT_SECRET to a random 64-char string
docker compose up --build -d

# Verify all containers are healthy
docker compose ps

# Wait for health checks to pass, then:
curl http://localhost:8080/api/v1/health
# → {"status":"ok","service":"VoxMeet API Gateway"}
```

---

## 1. Infrastructure Smoke Tests

| # | Test | Steps | Expected |
|---|---|---|---|
| 1.1 | PostgreSQL | `docker compose exec postgres psql -U postgres -d voxmeet -c "\dt"` | 5 tables: `messages`, `room_members`, `rooms`, `sessions`, `users` |
| 1.2 | RabbitMQ | Open `http://localhost:15672` login `guest:guest` | Management UI loads, no alarms |
| 1.3 | API Gateway | `curl -s http://localhost:8080/api/v1/health \| jq .` | `{"status":"ok","service":"VoxMeet API Gateway"}` |
| 1.4 | OpenAPI docs | Open `http://localhost:8080/api/v1/docs` in browser | Swagger UI loads with all endpoints |
| 1.5 | All containers | `docker compose ps --format "table {{.Name}}\t{{.Status}}"` | All services show `Up` or `healthy` |

---

## 2. Auth Flow

### 2.1 Registration

| # | Test | Steps | Expected |
|---|---|---|---|
| 2.1.1 | Successful registration | `curl -XPOST localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d '{"username":"alice","email":"alice@test.com","password":"password123"}'` | 201 with `token`, `user_id`, `username`, `email` |
| 2.1.2 | Duplicate email | Same request again | 409 `"email already registered"` |
| 2.1.3 | Duplicate username | `curl -XPOST ... -d '{"username":"alice","email":"bob@test.com","password":"password123"}'` | 409 `"username already taken"` |
| 2.1.4 | Short password | `curl -XPOST ... -d '{"username":"x","email":"x@t.com","password":"123"}'` | 400 `"password must be at least 6 characters"` |
| 2.1.5 | Invalid email | `curl -XPOST ... -d '{"username":"x","email":"notanemail","password":"password123"}'` | 400 `"invalid email address"` |
| 2.1.6 | Missing fields | `curl -XPOST ... -d '{}'` | 400 |

### 2.2 Login

| # | Test | Steps | Expected |
|---|---|---|---|
| 2.2.1 | Successful login | `curl -XPOST localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"alice@test.com","password":"password123"}'` | 200 with valid JWT `token` |
| 2.2.2 | Wrong password | `curl -XPOST ... -d '{"email":"alice@test.com","password":"wrong"}'` | 401 `"invalid email or password"` |
| 2.2.3 | Nonexistent user | `curl -XPOST ... -d '{"email":"noone@test.com","password":"x"}'` | 401 `"invalid email or password"` |

### 2.3 Token Validation

| # | Test | Steps | Expected |
|---|---|---|---|
| 2.3.1 | Valid token | `curl localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"` | 200 with user profile |
| 2.3.2 | Missing token | `curl localhost:8080/api/v1/me` | 401 |
| 2.3.3 | Invalid token | `curl localhost:8080/api/v1/me -H "Authorization: Bearer invalid"` | 401 |

---

## 3. Room Management

| # | Test | Steps | Expected |
|---|---|---|---|
| 3.1 | List rooms (empty) | `curl localhost:8080/api/v1/rooms -H "Authorization: Bearer $TOKEN"` | `[]` |
| 3.2 | Create room | `curl -XPOST localhost:8080/api/v1/rooms -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"name":"Test Room","is_public":true}'` | 201 with room `id`, `name`, `owner_id` |
| 3.3 | List rooms (1) | `curl localhost:8080/api/v1/rooms -H "Authorization: Bearer $TOKEN"` | Array with 1 room |
| 3.4 | Get room by ID | `curl localhost:8080/api/v1/rooms/$ROOM_ID -H "Authorization: Bearer $TOKEN"` | 200 with room details |
| 3.5 | Get nonexistent room | `curl localhost:8080/api/v1/rooms/00000000-0000-0000-0000-000000000000 -H "Authorization: Bearer $TOKEN"` | 404 |
| 3.6 | Update room name | `curl -XPUT localhost:8080/api/v1/rooms/$ROOM_ID -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"name":"Updated Room","is_public":false}'` | 200 with updated name |
| 3.7 | Delete room | `curl -XDELETE localhost:8080/api/v1/rooms/$ROOM_ID -H "Authorization: Bearer $TOKEN"` | 204 |
| 3.8 | Get deleted room | `curl localhost:8080/api/v1/rooms/$ROOM_ID -H "Authorization: Bearer $TOKEN"` | 404 |
| 3.9 | Create room (no auth) | `curl -XPOST localhost:8080/api/v1/rooms -d '{"name":"x"}'` | 401 |

---

## 4. WebSocket Signaling

### 4.1 Connection

| # | Test | Steps | Expected |
|---|---|---|---|
| 4.1.1 | Connect with valid token | Use a WebSocket client (e.g. `wscat`, Postman, or browser console) to connect: `ws://localhost:8080/api/v1/ws?token=$TOKEN` | Receives `{"type":"authenticated","user_id":"..."}` |
| 4.1.2 | Connect without token | `ws://localhost:8080/api/v1/ws` | Connection rejected with 401 |
| 4.1.3 | Connect with invalid token | `ws://localhost:8080/api/v1/ws?token=bad` | Connection rejected with 401 |
| 4.1.4 | Ping/pong | Send `{"type":"ping"}` | Receives `{"type":"pong"}` |

### 4.2 Join Room (Single Client)

| # | Test | Steps | Expected |
|---|---|---|---|
| 4.2.1 | Join room | Create room via REST, then send over WS: `{"type":"join_room","room_id":"$ROOM_ID"}` | Receives `sdp_offer` with SDP, then `room_state` with peers |
| 4.2.2 | Unknown message type | Send `{"type":"garbage"}` | Receives `{"type":"error","code":"unknown_type","message":"..."}` |
| 4.2.3 | Malformed JSON | Send `not json` | Receives `{"type":"error","code":"invalid_message","message":"..."}` |

### 4.3 Two-Peer Call

| Step | Client A (Alice) | Client B (Bob) | Expected |
|---|---|---|---|
| 1 | `{join_room, room-1}` via WS | — | Alice gets `sdp_offer` |
| 2 | — | `{join_room, room-1}` via WS | Bob gets `sdp_offer`, Alice gets `peer_joined` |
| 3 | Alice answers with `{sdp_answer, ..}` | — | ICE candidates flow |
| 4 | — | Bob answers with `{sdp_answer, ..}` | ICE candidates flow |
| 5 | Both clients should see `ice_candidate` events | | |
| 6 | Both clients should get `new_track` when the other publishes media | | |
| 7 | Send `{leave_room, room-1}` from Alice | Bob gets `peer_left` | |
| 8 | Bob sends `{leave_room, room-1}` | Room is cleaned up | |

### 4.4 ICE Candidate Flow

| # | Test | Steps | Expected |
|---|---|---|---|
| 4.4.1 | Send ICE candidate | `{"type":"ice_candidate","room_id":"$ROOM_ID","data":{"candidate":"candidate:...","sdp_mid":"0","sdp_mline_index":0}}` | SFU processes and sends ICE candidates to other peer |

---

## 5. Chat

| # | Test | Steps | Expected |
|---|---|---|---|
| 5.1 | Send chat message | Open 2 WS clients in same room. Send: `{"type":"chat.send","room_id":"$ROOM_ID","data":{"content":"Hello!"}}` | Both clients receive `{"type":"chat.message","content":"Hello!",...}` |
| 5.2 | Empty content | Send `{"type":"chat.send","room_id":"$ROOM_ID","data":{"content":""}}` | Message is rejected, no broadcast |
| 5.3 | Cross-room isolation | Send a chat from room-1. Client in room-2 should NOT receive it. | Correct — messages scoped to room |

---

## 6. Web Frontend (Browser)

| # | Test | Steps | Expected |
|---|---|---|---|
| 6.1 | Open app | Open `http://localhost:8080` (or `http://localhost:5173` for dev) | Redirected to `/login` |
| 6.2 | Register | Fill form, submit | Redirected to dashboard |
| 6.3 | Dashboard | See room list (empty) + "New Room" button | Dashboard renders |
| 6.4 | Create room | Click "New Room", type name, click Create | Redirected to room page |
| 6.5 | Room page | Camera/mic prompt appears | Allow — video tile shows "You" |
| 6.6 | Join from 2nd browser | Open incognito, register 2nd user, join same room | Both see each other's video |
| 6.7 | Mute/unmute | Click mic button | Icon toggles, audio stops/starts |
| 6.8 | Camera on/off | Click camera button | Icon toggles, video feed stops/starts |
| 6.9 | Screen share | Click screen share button | Browser shows screen picker, tile appears |
| 6.10 | Chat panel | Type message, hit Enter | Message appears in both clients |
| 6.11 | Leave room | Click Leave | Returns to dashboard |
| 6.12 | 404 page | Navigate to `http://localhost:5173/nonexistent` | "404 — Page not found" |

---

## 7. Multi-User Call (3+ Participants)

| # | Test | Steps | Expected |
|---|---|---|---|
| 7.1 | Alice, Bob, Charlie join same room | Open 3 browser windows/tabs, register 3 users, join same room | All 3 appear in video grid |
| 7.2 | Media flows to all | Each client's audio/video is visible/audible to others | All tiles show video |
| 7.3 | One leaves | Charlie leaves | Alice and Bob get `peer_left`, Charlie's tile disappears |
| 7.4 | Room empty | All leave | Room cleaned up — next join creates fresh room |

---

## 8. Error & Edge Cases

| # | Test | Steps | Expected |
|---|---|---|---|
| 8.1 | Camera denied | Click "Block" on camera prompt | App still works, shows "No Video" |
| 8.2 | Reconnect | Kill WiFi for 5s, restore | WebSocket reconnects, call recovers (or graceful error) |
| 8.3 | Register same email 2nd user after restart | Docker compose down, up, register same email | 409 Conflict |
| 8.4 | Large payload | Send 100KB chat message | Message is stored and broadcast |
| 8.5 | Rapid join/leave | Rapidly join/leave room 10 times | Room/peer maps cleanup correctly, no goroutine leaks |

---

## 9. Performance (Optional)

| # | Test | Steps | Expected |
|---|---|---|---|
| 9.1 | Memory usage | `docker stats` during 2-peer call | Each SFU session < 50MB |
| 9.2 | RTT latency | Measure WS ping/pong round trip | < 50ms locally |

---

## Test Results Log

| Date | Tester | Tests Passed | Tests Failed | Notes |
|---|---|---|---|---|
| | | / | / | |
| | | / | / | |
