# ConnectX API Summary

This document provides a summary of the WebSocket API for the ConnectX game server. It is intended for use by front-end developers to build a client UI.

## 1. WebSocket Connection

- **Endpoint**: `ws://<your_server_address>/ws`
- **Authentication**: The server expects a `token` cookie to be sent with the upgrade request. The value of the cookie is used as the `userID`.
  - **Example Header**: `Cookie: token=your_user_id_here`

## 2. Communication Protocol

All communication happens via binary WebSocket messages. The messages are JSON objects with a common structure for requests and responses.

### 2.1. Request Format

All client-to-server messages must follow this structure:

```json
{
  "type": 0,
  "id": "client-generated-request-id",
  "body": {
    "...": "..."
  }
}
```

- `type` (integer): The type of action to perform. See [Message Types](#3-message-types).
- `id` (string): A unique client-generated ID for the request. The server will use this in its response.
- `body` (object): A JSON object containing the payload for the specific action.

### 2.2. Response Format

All server-to-client messages follow this structure:

```json
{
  "req_id": "client-generated-request-id",
  "status": 0,
  "body": {
    "...": "..."
  }
}
```

- `req_id` (string): Corresponds to the `id` of the client's request. For server-pushed events (like an opponent's move), this may be `"-1"`.
- `status` (integer): A code indicating the result of the operation. See [Status Codes](#4-status-codes).
- `body` (object): A JSON object containing the response data. For errors, this will typically be `{"error": "message"}`.

---

## 3. Message Types (`type`)

These are the values for the `type` field in a `WsRequest`.

| Value | Constant Name                 | Description                               |
| :---- | :---------------------------- | :---------------------------------------- |
| `0`   | `MESSAGE_TYPE_REGISTER_MOVE_2D` | Submits a move for the current player.    |
| `1`   | `MESSAGE_TYPE_JOIN_MATCH_2D`    | Joins an existing 2D match.               |
| `2`   | `MESSAGE_TYPE_CREATE_MATCH_2D`  | Creates a new 2D match.                   |
| `3`   | `MESSAGE_TYPE_ABANDON_MATCH_2D` | (Not yet implemented) Abandons a match.   |
| `4`   | `MESSAGE_TYPE_ASK_DRAW_2D`      | (Not yet implemented) Proposes a draw.    |

## 4. Status Codes (`status`)

These are the values for the `status` field in a `WsResponse`.

| Value | Constant Name             | Description                                                              |
| :---- | :------------------------ | :----------------------------------------------------------------------- |
| `0`   | `WS_STATUS_OK`            | The request was successful.                                              |
| `1`   | `WS_STATUS_BAD_REQUEST`   | The request was malformed or invalid.                                    |
| `2`   | `WS_STATUS_SERVER_ERROR`  | An internal server error occurred.                                       |
| `3`   | `WS_STATUS_UNJOINABLE`    | The match cannot be joined (e.g., it's full or doesn't exist).           |
| `4`   | `WS_STATUS_ENEMY_JOINED`  | A server-pushed event indicating the opponent has joined the match.      |
| `5`   | `WS_STATUS_ENEMY_SENT_MOVE` | A server-pushed event indicating the opponent has made a move.           |
| `6`   | `WS_STATUS_GAMEOVER_WON`  | The game is over and the current player won.                             |
| `7`   | `WS_STATUS_GAMEOVER_LOST` | The game is over and the current player lost.                            |
| `8`   | `WS_STATUS_GAMEOVER_DRAW` | The game is over and it was a draw.                                      |

---

## 5. API Actions (2D Match)

### 5.1. Create Match

- **`type`**: `2` (`MESSAGE_TYPE_CREATE_MATCH_2D`)
- **Request Body**: `MatchOpts` object.
  ```json
  {
    "w": 7,
    "h": 6,
    "a": 4,
    "starts1": true,
    "t0": 60,
    "td": 0
  }
  ```
- **Success Response (`WS_STATUS_OK`)**:
  - **Body**: `{"id": "new-match-id"}`

### 5.2. Join Match

- **`type`**: `1` (`MESSAGE_TYPE_JOIN_MATCH_2D`)
- **Request Body**:
  ```json
  {
    "match_id": "existing-match-id"
  }
  ```
- **Success Response (`WS_STATUS_OK`)**:
  - **Body**: `Match2D` object (see [Data Models](#6-data-models)).
- **Notifications**:
  - When the second player joins, the first player will receive a `WS_STATUS_ENEMY_JOINED` message.
  - **Body**: `PlayerDTO` object of the player who just joined.

### 5.3. Register Move

- **`type`**: `0` (`MESSAGE_TYPE_REGISTER_MOVE_2D`)
- **Request Body**:
  ```json
  {
    "match_id": "existing-match-id",
    "col": 3,
    "sent_at": "2025-08-01T12:00:00Z"
  }
  ```
- **Success Response (`WS_STATUS_OK`)**:
  - **Body**: `null` (for a normal, non-game-ending move).
- **Game Over Responses**:
  - `WS_STATUS_GAMEOVER_WON`: The move resulted in a win.
    - **Body**: `{ "col": 3, "lines": [[...]], "time_left_p1": 55, "time_left_p2": 58 }`
  - `WS_STATUS_GAMEOVER_DRAW`: The move resulted in a draw.
    - **Body**: `{ "col": 3, "time_left_p1": 55, "time_left_p2": 58 }`
- **Notifications**:
  - The opponent will receive a `WS_STATUS_ENEMY_SENT_MOVE` message.
    - **Body**: `{ "col": 3, "time_left_p1": 55, "time_left_p2": 58 }`
  - If the move ends the game, the opponent will receive `WS_STATUS_GAMEOVER_LOST` or `WS_STATUS_GAMEOVER_DRAW`.

---

## 6. Data Models (JSON Structures)

### MatchOpts

Options for creating a match.

```json
{
  "w": 7,       // Width of the board (3-15)
  "h": 6,       // Height of the board (3-15)
  "a": 4,       // Number of pieces in a row to win (3-15)
  "starts1": true, // Does player 1 start?
  "t0": 60,     // Initial time for each player (seconds)
  "td": 0       // Time delta per move (not implemented)
}
```

### PlayerDTO

Public data for a player.

```json
{
  "id": "player-id",
  "TimeLeft": 60,
  "Nick": "PlayerNickname",
  "ImgURL": "http://example.com/avatar.png"
}
```

### Match2D

The full state of a 2D match.

```json
{
  "Board": [
    [0, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 1, 0, 0, 0],
    [0, 0, 2, 1, 0, 0, 0],
    [1, 2, 1, 2, 0, 0, 0]
  ],
  "P1": { "ID": "player1-id", "TimeLeft": 55 },
  "P2": { "ID": "player2-id", "TimeLeft": 58 },
  "Opts": { "...": "..." }, // MatchOpts object
  "Moves": [
    { "Col": 2, "RegisteredAt": "..." },
    { "Col": 1, "RegisteredAt": "..." }
  ],
  "StartedAt": "2025-08-01T11:59:00Z"
}
```
- **Board Slots**: `0` = Empty, `1` = Player 1, `2` = Player 2.
