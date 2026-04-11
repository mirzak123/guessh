# GueSSH

Welcome to GueSSH, a head-to-head multiplayer game inspired by Wordle, available over SSH.

![demo](./assets/demo.gif)

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

This will spin up two containers, one for the game server, and another for the SSH server.
For a more detailed explanation on the project components, check out the section on [implementation details](#implementation-details).

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
make run-gamed      # C server that handles all of the game logic

make run-sshd       # Run the SSH server, which serves the TUI

make run-tui        # Run the TUI that connects to the game server

make run-cli        # Run the CLI that connects to the game server

make test-gamed     # Run game server tests
```

If using this method, you don't need to spin up the SSH server at all, as `make run-tui` will give you the
TUI and connect to the game server without needing to go through the SSH server. You do, however, need to
run `make run-server`, before trying to connect to the game server (obviously).

In case of any port issues, you can check out the [environment variables](#environment-variables) section to check port defaults, and
adjust the ports, if the defaults are occupied on your machine.

## Implementation Details

The project is composed of 4 components:

### guessh-gamed

TCP server written in C that handles all game logic. It communicated with clients (guessh-ssh, guessh-tui, or guessh-cli) via a JSON messages
encoded over binary, using cJSON, over a raw TCP stream. No predefined application level protocol is used.
Messages are prepended with a 4 byte length prefix to separate TCP segments in case of sticking.
The communication between the game server and clients is explained in more detail in the [Client-Server Protocol](#client-server-protocol) section.

It's single threaded and used poll(2) for handling client file descriptors.

It has supports timers for limiting turn time limits.
It doesn't use POSIX timers, but instead handles a linked list of sorted timers and a timeout on the poll(2) call to check them periodically.

### guessh-sshd

SSH server that serves the TUI over SSH, and connects to the game server via TCP.
It's written in Go, using [charmbracelet/wish](https://github.com/charmbracelet/wish) to serve the TUI over an SSH connection.

### guessh-tui

Terminal User Interface (TUI) that connects directly to the game server, without going through SSH.
It's written in Go, using [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea).
This is a simple component that just wraps the bubbletea model, and opens a TCP connection directly to the game server.
The same bubbletea model is served by the SSH server.
`guessh-tui` is made to simplify development, so that I don't have to spin up the SSH server all the time.

### guessh-cli

This is a simple TCP client that connects to the game server. Like other client components, it's written in Go.
It's basically just netcat with the addition of handling length prefixes,
so that the server doesn't interpret messages as junk, and drop the connection.

It's used for debugging, and checking the server statistics while it's running.

## Environment Variables

The following environment variables are used by one or multiple project components:

| Variable           | Default         | Description                                                 |
| ------------------ | --------------- | ----------------------------------------------------------- |
| GUESSH_SERVER_PORT | 2480            | Port that `guessh-gamed` listens on.                        |
| GAME_SERVER_ADDR   | localhost:2480  | Address the client components dial to reach `gusssh-gamed`. |
| WORDS_PATH         | ./words         | Directory containing the .txt word lists.                   |
| GUESSH_SSH_ADDR    | :2222           | Address that `guessh-sshd` listens on.                      |
| HOST_KEY_PATH      | .ssh/id_ed25519 | Path to the SSH Server's persistent identity key.           |
| LOG_LEVEL          | INFO            | Logging verbosity (DEBUG \| INFO \| ERROR).                 |

## Client-Server Protocol

The game server (`guessh-gamed`) communicates with clients over a TCP connection, without a common application level protocol,
like HTTP or WebSocket, to handle semantics. Instead, it uses the TCP connection as a bidirectional stream, with
no concepts of a request-response cycle.

Both the client and server communicate events encoded as JSON messages. Each message contains a `type` field.
Each event sent by the client has a corresponding handler on the server, and vice-versa.

### TCP Stream Fragmentation

While TCP guarantees that segments, will arrive in order, it does not guarantee that one read(2) call
on the reader's side of the socket, will consume exactly the same number of bytes sent by a single
write(2) call on the writer's side. This means that without a way to fragment the TCP stream, two JSON
events can arrive stuck together, or a single JSON event could be split into two read(2) calls.

To tackle this, a length prefix of 4 bytes is used before each message. Before sending an event,
the sender will calculate the lenght of the payload, and send it in a 4-byte long, big endian encoded message.
The listener first reads exactly 4 bytes to understand how big of a payload to read afterwards, and then it
reads bytes from the TCP stream until all of the payload is consumed, which is always a single JSON event.

### Events

Events are split into client and server events, depending on the sender.

#### Client Events

| Type             | Content                                                                                                                                          | Additional Info                              |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------- |
| CREATE_MATCH     | {"type": "CREATE_MATCH", "mode": single, "format": string, "rounds": number, "wordLength": number, "turnTimeout": number, "playerName"?: string} | playerName required only in multiplayer mode |
| JOIN_ROOM        | {"type": "JOIN_ROOM", "roomId": string, "playerName": string}                                                                                    |                                              |
| MAKE_GUESS       | {"type": "MAKE_GUESS", "guess": string}                                                                                                          |                                              |
| REQUEST_REMATCH  | {"type": "REQUEST_REMATCH"}                                                                                                                      |                                              |
| DENY_REMATCH     | {"type": "DENY_REMATCH"}                                                                                                                         |                                              |
| TYPING           | {"type": "TYPING", "value": string}                                                                                                              |                                              |
| LEAVE_MATCH      | {"type": "LEAVE_MATCH"}                                                                                                                          |                                              |
| READY_NEXT_ROUND | {"type": "READY_NEXT_ROUND"}                                                                                                                     |                                              |
| SHOW_STATS       | {"type": "SHOW_STATS"}                                                                                                                           |                                              |

#### Server Events

| Type                    | Content                                                                                                                                                | Additional Info                                    |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------- |
| ERROR                   | {"type": "ERROR", "reason": string}                                                                                                                    |                                                    |
| ROOM_CREATED            | {"type": "ROOM_CREATED", "roomId": string}                                                                                                             |                                                    |
| ROOM_JOINED             | {"type": "ROOM_JOINED", "roomId": string}                                                                                                              |                                                    |
| ROOM_JOIN_FAILED        | {"type": "ROOM_JOIN_FAILED", "roomId": string, "reason": string}                                                                                       | Room full or invalid room ID                       |
| WAIT_OPPONENT_JOIN      | {"type": "WAIT_OPPONENT_JOIN"}                                                                                                                         |                                                    |
| OPPONENT_DENIED_REMATCH | {"type": "OPPONENT_DENIED_REMATCH"}                                                                                                                    |                                                    |
| OPPONENT_LEFT           | {"type": "OPPONENT_LEFT"}                                                                                                                              |                                                    |
| MATCH_STARTED           | {"type": "MATCH_STARTED", "matchId": string, "format": string, "rounds": number, "wordLength": number, "turnTimeout": number, "opponentName"?: string} | opponentName is only included in MULTI_REMOTE mode |
| ROUND_STARTED           | {"type": "ROUND_STARTED", "roundNumber": number, "maxAttempts": number}                                                                                |                                                    |
| WAIT_GUESS              | {"type": "WAIT_GUESS"}                                                                                                                                 |                                                    |
| WAIT_OPPONENT_GUESS     | {"type": "WAIT_OPPONENT_GUESS"}                                                                                                                        |                                                    |
| GUESS_RESULT            | {"type": "GUESS_RESULT", "guess": string, "feedback": number[]\[]}                                                                                     |                                                    |
| ROUND_FINISHED          | {"type": "ROUND_FINISHED", "points": number, "word": string, "postRoundTimeout": number}                                                               |                                                    |
| MATCH_FINISHED          | {"type": "MATCH_FINISHED", outcome: number , "opponentLeft": boolean}                                                                                  |                                                    |
| OPPONENT_TYPING         | {"type": "OPPONENT_TYPING", "value": string}                                                                                                           |                                                    |
| STATS                   | {...}                                                                                                                                                  | Game server statistics                             |
