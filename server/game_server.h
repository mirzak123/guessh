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

#define E_MALFORMED_MESSAGE "Malformed message received"
#define E_UNSUPPORTED_MESSAGE_TYPE "Unsupported message type"
#define E_NOT_IMPLEMENTED "Not Implemented"
#define E_MISSING_FIELD(field) "Missing '" field "' field"
#define E_INVALID_TYPE(field, expected_type) "Invalid type of '" field "' field, expected: " expected_type
#define E_INVALID_ROUNDS "Round number must be between 1 and " STR(MAX_ROUNDS)
#define E_UNSUPPORTED_MODE "Unsupported mode"

typedef enum {
  MALFORMED_MESSAGE = -1,
  UNSUPPORTED_MESSAGE_TYPE,

  // Client
  CREATE_ROOM,
  CREATE_MATCH,
  JOIN_ROOM,
  MAKE_GUESS,
  REQUEST_REMATCH,
  LEAVE_MATCH,

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
  GUESS_RESULT,
  ROUND_FINISHED,
  MATCH_FINISHED,
  BYE,
} MessageType;

typedef struct {
  Match *matches[MAX_MATCHES];
  int match_number;
} GameServer;

GameServer *GS_create(void);
void GS_handle_request(GameServer *gs, int client_fd, char *data, size_t size);
long GS_create_match(GameServer *gs, int client_fd, int nrounds, char *err);
Match *GS_get_match_by_player(GameServer *gs, int player_fd);
void GS_destroy(GameServer *gs);
MessageType GS_parse_message(char *data, size_t size, cJSON **out);

// message send
void GS_send_json(int client_fd, cJSON *json);
void GS_send_error(int client_fd, char *reason);

// message handlers
void GS_handle_create_match(GameServer *gs, int client_fd, cJSON *json_request);

#endif // !GAME_SERVER_H
