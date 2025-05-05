# Welcome to GueSSH

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

| #   | Name                      | Sent by: C or S | Description                                                                    |
| --- | ------------------------- | --------------- | ------------------------------------------------------------------------------ |
| 0   | CONNECTED                 | S               |                                                                                |
| 1   | ROOM_CREATE               | C               |                                                                                |
| 2   | START_SINGLE_PLAYER_MATCH | C               | Client tells server to start a match without creating room                     |
| 3   | ROOM_CREATED              | S               |                                                                                |
| 4   | ROOM_JOIN                 | C               |                                                                                |
| 5   | ROOM_JOINED               | S               |                                                                                |
| 6   | WAIT_OPPONENT_JOIN        | S               | One player created a room and is waiting for the other player to join the room |
| 7   | MATCH_STARTED             | S               |                                                                                |
| 8   | ROUND_STARTED             | S               |                                                                                |
| 9   | WAIT_GUESS                | S               |                                                                                |
| 10  | WAIT_OPPONENT_GUESS       | S               |                                                                                |
| 11  | MAKE_GUESS                | C               |                                                                                |
| 12  | GUESS_OUTCOME             | S               |                                                                                |
| 13  | ROUND_FINISHED            | S               |                                                                                |
| 14  | REQUEST_REMATCH           | C               |                                                                                |
| 15  | MATCH_FINISHED            | S               |                                                                                |
| 16  | DISCONNECTING             | S               |                                                                                |
