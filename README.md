# GueSSH

Welcome to GueSSH, a wordle-like head-to-head multiplayer game, available over SSH.

## Server

### File structure

- main.c - main loop
- game.c - game logic
- room.c - creating, joining and deleting rooms
- network.c - socket setup for communication with cli

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
              - Create new room 5. How many rounds? (number input)

## Client-Server Protocol

| #   | Name                | Sent by: C or S | Description                                                                    |
| --- | ------------------- | --------------- | ------------------------------------------------------------------------------ |
| 0   | CREATE_ROOM         | C               |                                                                                |
|     | CONNECTED           | S               |                                                                                |
|     | CREATE_MATCH        | C               | Client tells server to start a match                                           |
|     | ROOM_CREATED        | S               |                                                                                |
|     | ROOM_JOIN           | C               |                                                                                |
|     | ROOM_JOINED         | S               |                                                                                |
|     | ROOM_JOIN_FAILED    | S               | Room full                                                                      |
|     | EXIT_ROOM           | S/C             | Potentially triggers server to kick other participant if match is over         |
|     | WAIT_OPPONENT_JOIN  | S               | One player created a room and is waiting for the other player to join the room |
|     | MATCH_STARTED       | S               |                                                                                |
|     | ROUND_STARTED       | S               |                                                                                |
|     | WAIT_GUESS          | S               |                                                                                |
|     | WAIT_OPPONENT_GUESS | S               |                                                                                |
|     | MAKE_GUESS          | C               |                                                                                |
|     | ACCEPTED_GUESS      | S               |                                                                                |
|     | INVALID_GUESS       | S               |                                                                                |
|     | ROUND_FINISHED      | S               |                                                                                |
|     | REQUEST_REMATCH     | C               |                                                                                |
|     | MATCH_FINISHED      | S               |                                                                                |
|     | DISCONNECTING       | S/C             |                                                                                |
