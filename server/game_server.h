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

#define STRING "string"
#define NUMBER "number"

#define E_MISSING_FIELD(field) "Missing '" field "' field"
#define E_INVALID_TYPE(field, expected_type) "Invalid type of '" field "' field, expected: " expected_type
#define E_INVALID_ROUNDS "Round number must be between 1 and " STR(MAX_ROUNDS)
#define E_UNSUPPORTED_MODE "Unsupported mode"

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
  ERROR,
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
// message send
void GS_send_json(int client_fd, cJSON *json);
void GS_send_error(int client_fd, char *reason);

#endif // !GAME_SERVER_H
