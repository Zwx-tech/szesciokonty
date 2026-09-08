# Hex

1v1 hex tactics web game. Frontend: PixiJS v8 + TypeScript. Backend: Go + WebSocket.

## Run

Terminal 1 — backend:

```bash
cd backend
go run ./cmd/server
```

Listens on `:3000` (`ADDR` overrides).

Terminal 2 — frontend:

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:8080. Vite proxies `/ws` and `/health` to the backend.

## Layout

- `frontend/` — Pixi client (`src/protocol.ts`, `src/net.ts`)
- `backend/` — game server (`internal/protocol` mirrors the client messages)

Wire format: JSON text frames, `{ "v": 1, "type": "..." , ... }`. Client sends intents; server pushes `room_state`, `match_state`, or `error`.
