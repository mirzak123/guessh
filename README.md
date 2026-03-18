# GueSSH

Welcome to GueSSH, a head-to-head multiplayer game inspired by Wordle, available over SSH.

## How to Play?

### Connect with SSH

Open a terminal and run the command below to play the game instantly (assuming you have an SSH client installed):

```bash
ssh guessh.duckdns.org
```

> **_NOTE:_** When playing for the first time you will likely be prompted to specify whether you trust this host and
> want to save its address to your known hosts file. You can safely type "yes" as this connection is harmless.

### Run Locally

You can run the game locally by running Docker containers with [Docker Compose](https://docs.docker.com/compose/), or
building and running the source using [make](https://www.gnu.org/software/make/). I recommend using Docker Compose, to mitigate any dependency issues.

#### Run with Docker Compose

The game comes with two `Dockerfile`s and a `docker-compose.yml` which can be used to run the game locally by running:

```bash
docker compose up --build

# or if you have docker-compose as a separate binary
docker-compose up --build
```

<!-- TODO: Link this to the game implementation section -->

This will spin up two containers, one for the game server, and another for the SSH server.

> **_NOTE_:** The SSH server will try to bind to port 22 on your machine, which is most likely taken.
> In that case, update the `ports` section in `docker-compose.yaml` to bind to a different port (e.g. `2222:2222`).

Then you can connect to the SSH server by running:

```bash
ssh localhost -p <PORT>
```

To stop the docker containers, run:

```bash
docker compose down

# or if you have docker-compose as a separate binary
docker-compose down
```

#### Run with `make`

You can alternatively build and run the source using `make`.

> **_NOTE_:** You will likely need to resolve dependencies on your own. Currently the only dependency that might
> require an install is `libcjson`, for the game server, but other dependencies might get used in the future.

With `make` you can run 3 different components (and tests, if you're into that):

```bash
make run-server     # C server that handles all of the game logic

make run-cli        # Run the TUI that connects to the game server

make run-ssh        # Run the SSH server, which serves the TUI

make run-tests      # Run game server tests
```

If using this method, you don't need to spin up the SSH server at all, as `make run-cli` will give you the
TUI and connect to the game server without needing to go through the SSH server. You do, however, need to
run `make run-server`, before trying to connect to the game server (obviously).

## Game Details

The game can be played in the following

## Environment Variables

| Variable           | Default         | Description                                          |
| ------------------ | --------------- | ---------------------------------------------------- |
| GUESSH_SERVER_PORT | 2480            | Port the C Game Server listens on.                   |
| GAME_SERVER_ADDR   | localhost:2480  | Address the Go Frontend dials to reach the C Server. |
| WORDS_PATH         | ./words         | Directory containing the .txt word lists.            |
| GUESSH_SSH_ADDR    | :2222           | Address the Go SSH Server listens on.                |
| HOST_KEY_PATH      | .ssh/id_ed25519 | Path to the SSH Server's persistent identity key.    |
| LOG_LEVEL          | INFO            | Logging verbosity (DEBUG \| INFO \| ERROR).          |

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

| Type            | Content                                                                                                                   | Additional Info                              |
| --------------- | ------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| CREATE_MATCH    | {"type": "CREATE_MATCH", "mode": single, "format": string, "rounds": number, "wordLength": number, "playerName"?: string} |                                              |
| JOIN_ROOM       | {"type": "JOIN_ROOM", "roomId": string, "playerName"?: string}                                                            | playerName required only in multiplayer mode |
| MAKE_GUESS      | {"type": "MAKE_GUESS", "guess": string}                                                                                   |                                              |
| REQUEST_REMATCH | {"type": "REQUEST_REMATCH"}                                                                                               |                                              |
| DENY_REMATCH    | {"type": "DENY_REMATCH"}                                                                                                  |                                              |
| TYPING          | {"type": "TYPING", "value": string}                                                                                       |                                              |
| LEAVE_MATCH     | {"type": "LEAVE_MATCH"}                                                                                                   |                                              |

### Server Message Types

| Type                    | Content                                                                                                                         | Additional Info                                    |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------- |
| ERROR                   | {"type": "ERROR", "reason": string}                                                                                             |                                                    |
| ROOM_CREATED            | {"type": "ROOM_CREATED", "roomId": string}                                                                                      |                                                    |
| ROOM_JOINED             | {"type": "ROOM_JOINED", "roomId": string}                                                                                       |                                                    |
| ROOM_JOIN_FAILED        | {"type": "ROOM_JOIN_FAILED", "roomId": string, "reason": string}                                                                | Room full or invalid room ID                       |
| WAIT_OPPONENT_JOIN      | {"type": "WAIT_OPPONENT_JOIN"}                                                                                                  |                                                    |
| OPPONENT_DENIED_REMATCH | {"type": "OPPONENT_DENIED_REMATCH"}                                                                                             |                                                    |
| OPPONENT_LEFT           | {"type": "OPPONENT_LEFT"}                                                                                                       |                                                    |
| MATCH_STARTED           | {"type": "MATCH_STARTED", "matchId": string, "format": string, "rounds": number, "wordLength": number, "opponentName"?: string} | opponentName is only included in MULTI_REMOTE mode |
| ROUND_STARTED           | {"type": "ROUND_STARTED", "roundNumber": number, "maxAttempts": number}                                                         |                                                    |
| WAIT_GUESS              | {"type": "WAIT_GUESS"}                                                                                                          |                                                    |
| WAIT_OPPONENT_GUESS     | {"type": "WAIT_OPPONENT_GUESS"}                                                                                                 |                                                    |
| GUESS_RESULT            | {"type": "GUESS_RESULT", "guess": string, "feedback": number[]\[]}                                                              |                                                    |
| ROUND_FINISHED          | {"type": "ROUND_FINISHED", "points": number, "word": string}                                                                    |                                                    |
| MATCH_FINISHED          | {"type": "MATCH_FINISHED", outcome: number , "opponentLeft": boolean}                                                           | outcome only relevant for multiplayer games        |
| OPPONENT_TYPING         | {"type": "OPPONENT_TYPING", "value": string}                                                                                    |
