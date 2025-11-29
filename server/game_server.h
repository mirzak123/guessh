#ifndef GAME_SERVER_H
#define GAME_SERVER_H

#include "game_types.h"
#include <cjson/cJSON.h>
#include <stddef.h>

#define MAX_MATCHES 10
#define MAX_ROUNDS 5

// don't ask...
#define STR_HELPER(x) #x
#define STR(x) STR_HELPER(x)

#define E_INVALID_ROUNDS "Round number must be between 1 and " STR(MAX_ROUNDS) "\n"

typedef enum {
  MALFORMED_MESSAGE = -1,
  UNSUPPORTED_MESSAGE_TYPE,

  // Both
  MATCH_INFO,
  ROUND_INFO,
  BYE,

  // Client
  CREATE_ROOM,
  CREATE_MATCH,
  CREATE_ROUND,
  JOIN_ROOM,
  MAKE_GUESS,
  REQUEST_REMATCH,
  EXIT_MATCH,

  // Server
  CONNECTED,
  ROOM_CREATED,
  ROOM_JOINED,
  ROOM_JOIN_FAILED,
  WAIT_OPPONENT_JOIN,
  MATCH_STARTED,
  ROUND_STARTED,
  WAIT_GUESS,
  WAIT_OPPONENT_GUESS,
  ACCEPTED_GUESS,
  INVALID_GUESS,
  ROUND_FINISHED,
  MATCH_FINISHED,
} MessageType;

typedef struct {
  Match *matches[MAX_MATCHES];
  int nmatches;
} GameServer;

GameServer *GS_create(void);
void GS_request(GameServer *gs, int client_fd, char *data, size_t size);
long GS_create_match(GameServer *gs, int client_fd, int nrounds, char *err);
Match *GS_get_match_by_player(GameServer *gs, int player_fd);
void GS_destroy(GameServer *gs);
MessageType GS_parse_message(char *data, size_t size, cJSON *out);

#endif // !GAME_SERVER_H
