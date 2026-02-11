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

1.  Are you lonely?
    - Single player
      1. How many rounds? (number input)
      2. How many letters? (number input)
    - Two player remote
      1. Session:
         - Join existing room
           1. Name (string input)
           2. Room key (string input)
         - Create new room
           1. Name (string input)
           2. How many rounds? (number input)
           3. How many letters? (number input)

## Client-Server Protocol

### Client Message Types

| Type            | Content                                                                            |
| --------------- | ---------------------------------------------------------------------------------- |
| CREATE_ROOM     | {"type": "CREATE_ROOM"}                                                            |
| CREATE_MATCH    | {"type": "CREATE_MATCH", "mode": "SINGLE", "rounds": number, "wordLength": number} |
| JOIN_ROOM       |                                                                                    |
| MAKE_GUESS      | {"type": "MAKE_GUESS", "guess": string}                                            |
| REQUEST_REMATCH |                                                                                    |
| LEAVE_MATCH     | {"type": "LEAVE_MATCH"}                                                            |

### Server Message Types

| Type                | Content                                                                              | Additional info                                                         |
| ------------------- | ------------------------------------------------------------------------------------ | ----------------------------------------------------------------------- |
| ERROR               | {"type": "ERROR", "reason": string}                                                  |                                                                         |
| CONNECTED           |                                                                                      |                                                                         |
| ROOM_CREATED        | {"type": "ROOM_CREATED", "roomId": string}                                           |                                                                         |
| ROOM_JOINED         | {"type": "ROOM_JOINED"}                                                              |                                                                         |
| ROOM_JOIN_FAILED    | {"type": "ROOM_JOIN_FAILED"}                                                         | Room full                                                               |
| WAIT_OPPONENT_JOIN  | {"type": "WAIT_OPPONENT_JOIN"}                                                       |                                                                         |
| MATCH_STARTED       | {"type": "MATCH_STARTED", "matchId": string, "rounds": number, "wordLength": number} |                                                                         |
| ROUND_STARTED       | {"type": "ROUND_STARTED", "roundNumber": number, "maxAttempts": number}              |                                                                         |
| WAIT_GUESS          | {"type": "WAIT_GUESS"}                                                               |                                                                         |
| WAIT_OPPONENT_GUESS | {"type": "WAIT_OPPONENT_GUESS"}                                                      |                                                                         |
| GUESS_RESULT        | {"type": "GUESS_RESULT", "success": boolean, "guess": string, "feedback": number[]}  |                                                                         |
| ROUND_FINISHED      | {"type": "ROUND_FINISHED", "success": boolean, "word": string}                       | In multiplayer we need to provide a field indicating the winning player |
| MATCH_FINISHED      | {"type": "MATCH_FINISHED", "winner": string}                                         | Winner only relevant for multiplayer games.                             |
| BYE                 | {"type": "BYE"}                                                                      |                                                                         |
