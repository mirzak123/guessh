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

| Type            | Content                                                                                                   | Additional Info                              |
| --------------- | --------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| CREATE_MATCH    | {"type": "CREATE_MATCH", "mode": "SINGLE", "rounds": number, "wordLength": number, "playerName"?: string} | playerName required only in multiplayer mode |
| JOIN_ROOM       | {"type": "JOIN_ROOM", "roomId": string, "playerName"?: string}                                            | playerName required only in multiplayer mode |
| MAKE_GUESS      | {"type": "MAKE_GUESS", "guess": string}                                                                   |                                              |
| REQUEST_REMATCH | {"type": "REQUEST_REMATCH"}                                                                               |                                              |
| DENY_REMATCH    | {"type": "DENY_REMATCH"}                                                                                  |                                              |
| TYPING          | {"type": "TYPING", "value": string}                                                                       |                                              |
| LEAVE_MATCH     | {"type": "LEAVE_MATCH"}                                                                                   |                                              |

### Server Message Types

| Type                    | Content                                                                                                       | Additional Info                                         |
| ----------------------- | ------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| ERROR                   | {"type": "ERROR", "reason": string}                                                                           |                                                         |
| ROOM_CREATED            | {"type": "ROOM_CREATED", "roomId": string}                                                                    |                                                         |
| ROOM_JOINED             | {"type": "ROOM_JOINED", "roomId": string}                                                                     |                                                         |
| ROOM_JOIN_FAILED        | {"type": "ROOM_JOIN_FAILED", "roomId": string, "reason": string}                                              | Room full or invalid room ID                            |
| WAIT_OPPONENT_JOIN      | {"type": "WAIT_OPPONENT_JOIN"}                                                                                |                                                         |
| OPPONENT_DENIED_REMATCH | {"type": "OPPONENT_DENIED_REMATCH"}                                                                           |                                                         |
| OPPONENT_LEFT           | {"type": "OPPONENT_LEFT"}                                                                                     |                                                         |
| MATCH_STARTED           | {"type": "MATCH_STARTED", "matchId": string, "rounds": number, "wordLength": number, "opponentName"?: string} | opponentName will not be included in single player mode |
| ROUND_STARTED           | {"type": "ROUND_STARTED", "roundNumber": number, "maxAttempts": number}                                       |                                                         |
| WAIT_GUESS              | {"type": "WAIT_GUESS"}                                                                                        |                                                         |
| WAIT_OPPONENT_GUESS     | {"type": "WAIT_OPPONENT_GUESS"}                                                                               |                                                         |
| GUESS_RESULT            | {"type": "GUESS_RESULT", "success": boolean, "guess": string, "feedback": number[]}                           |                                                         |
| ROUND_FINISHED          | {"type": "ROUND_FINISHED", "outcome": number, "word": string}                                                 |                                                         |
| MATCH_FINISHED          | {"type": "MATCH_FINISHED", "outcome": number, "opponentLeft": boolean}                                        | outcome only relevant for multiplayer games.            |
| OPPONENT_TYPING         | {"type": "OPPONENT_TYPING", "value": string}                                                                  |
