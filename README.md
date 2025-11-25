# GueSSH

Welcome to GueSSH, a wordle-like head-to-head multiplayer game, available over SSH.

## Server

### File structure

- main.c - main loop
- game.c - game logic
- room.c - creating, joining and deleting rooms
- network.c - socket setup for communication with client

### Game Logic

...

## CLI

First checks if server is up and running on localhost:port. If not, display message that the server is not running.

### Prompts

---

**TODO**

Find a better way to represent flow below 🙏🙏🙏

---

1.  Choose game mode
    - 1 word
    - 4 words
2.  Are you lonely?
    - Single player
      1. How many rounds? (number input)
    - Multi player
      1. Choose mode
         - Local (same terminal session)
           1. How many rounds? (number input)
         - Remote
           1. Session:
              - Join existing room (room code input)
              - Create new room
                1. How many rounds? (number input)

## Client-Server Protocol

### Shared Message Types (Client & Server)

| Type       | Description                                           |
| ---------- | ----------------------------------------------------- |
| MATCH_INFO |                                                       |
| ROUND_INFO |                                                       |
| BYE        | Potentially triggers server to kick other participant |

### Client Message Types

| Type            | Description                          |
| --------------- | ------------------------------------ |
| CREATE_ROOM     |                                      |
| CREATE_MATCH    | Client tells server to start a match |
| ROOM_JOIN       |                                      |
| MAKE_GUESS      |                                      |
| REQUEST_REMATCH |                                      |
| EXIT_ROOM       |                                      |

### Server Message Types

| Type                | Description                                                                    |
| ------------------- | ------------------------------------------------------------------------------ |
| CONNECTED           |                                                                                |
| ROOM_CREATED        |                                                                                |
| ROOM_JOINED         |                                                                                |
| ROOM_JOIN_FAILED    | Room full                                                                      |
| WAIT_OPPONENT_JOIN  | One player created a room and is waiting for the other player to join the room |
| MATCH_STARTED       |                                                                                |
| ROUND_STARTED       |                                                                                |
| WAIT_GUESS          |                                                                                |
| WAIT_OPPONENT_GUESS |                                                                                |
| ACCEPTED_GUESS      |                                                                                |
| INVALID_GUESS       |                                                                                |
| ROUND_FINISHED      |                                                                                |
| MATCH_FINISHED      |                                                                                |
| MATCH_INFO          |                                                                                |
| ROUND_INFO          |                                                                                |
