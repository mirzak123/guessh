#include "game_server.h"
#include "game_logic.h"
#include "game_types.h"
#include "json_messages.h"
#include <_string.h>
#include <cjson/cJSON.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <time.h>

GameServer *GS_create(void) {
  GameServer *gs;

  gs = malloc(sizeof(GameServer));
  if (!gs) {
    perror("malloc");
    exit(1);
  }

  return gs;
}

MessageType GS_parse_message(char *data, size_t size, cJSON **json_out) {
  cJSON *json_type = NULL;
  char *type;
  MessageType mt;

  *json_out = cJSON_ParseWithLength(data, size);
  if (*json_out == NULL) {
    printf("[GS_parse_message] cJSON failed to parse message\n");
    return MALFORMED_MESSAGE;
  }
  printf("[GS_parse_message] json_out: %s\n", cJSON_PrintUnformatted(*json_out));

  json_type = cJSON_GetObjectItem(*json_out, "type");
  if (json_type == NULL) {
    printf("[GS_parse_message] message missing 'type' field\n");
    return MALFORMED_MESSAGE;
  }

  if (!cJSON_IsString(json_type)) {
    printf("[GS_parse_message] message 'type' field is not a string\n");
    return MALFORMED_MESSAGE;
  }
  type = cJSON_GetStringValue(json_type);

  if (!strcmp("BYE", type)) {
    mt = BYE;
  } else if (!strcmp("CREATE_ROOM", type)) {
    mt = CREATE_ROOM;
  } else if (!strcmp("CREATE_MATCH", type)) {
    mt = CREATE_MATCH;
  } else if (!strcmp("JOIN_ROOM", type)) {
    mt = JOIN_ROOM;
  } else if (!strcmp("MAKE_GUESS", type)) {
    mt = MAKE_GUESS;
  } else if (!strcmp("REQUEST_REMATCH", type)) {
    mt = REQUEST_REMATCH;
  } else if (!strcmp("LEAVE_MATCH", type)) {
    mt = LEAVE_MATCH;
  } else if (!strcmp("CONNECTED", type)) {
    mt = CONNECTED;
  } else if (!strcmp("ROOM_CREATED", type)) {
    mt = ROOM_CREATED;
  } else if (!strcmp("ROOM_JOINED", type)) {
    mt = ROOM_JOINED;
  } else if (!strcmp("ROOM_JOIN_FAILED", type)) {
    mt = ROOM_JOIN_FAILED;
  } else if (!strcmp("WAIT_OPPONENT_JOIN", type)) {
    mt = WAIT_OPPONENT_JOIN;
  } else if (!strcmp("MATCH_STARTED", type)) {
    mt = MATCH_STARTED;
  } else if (!strcmp("ROUND_STARTED", type)) {
    mt = ROUND_STARTED;
  } else if (!strcmp("WAIT_GUESS", type)) {
    mt = WAIT_GUESS;
  } else if (!strcmp("WAIT_OPPONENT_GUESS", type)) {
    mt = WAIT_OPPONENT_GUESS;
  } else if (!strcmp("GUESS_RESULT", type)) {
    mt = GUESS_RESULT;
  } else if (!strcmp("ROUND_FINISHED", type)) {
    mt = ROUND_FINISHED;
  } else if (!strcmp("MATCH_FINISHED", type)) {
    mt = MATCH_FINISHED;
  } else {
    mt = UNSUPPORTED_MESSAGE_TYPE;
  }

  return mt;
}

void GS_handle_request(GameServer *gs, int client_fd, char *data, size_t size) {
  cJSON *json_request = NULL;
  MessageType mt;

  mt = GS_parse_message(data, size, &json_request);

  switch (mt) {
  case MALFORMED_MESSAGE:
    printf("[GS_handle_request] Malformed request received\n");
    if (json_request) {
      cJSON_Delete(json_request);
    }
    GS_send_error(client_fd, E_MALFORMED_MESSAGE);
    return;
  case CREATE_MATCH:
    GS_handle_create_match(gs, client_fd, json_request);
    break;
  case CREATE_ROOM:
    GS_send_error(client_fd, E_NOT_IMPLEMENTED);
    break;
  case JOIN_ROOM:
    GS_send_error(client_fd, E_NOT_IMPLEMENTED);
    break;
  case MAKE_GUESS:
    GS_send_error(client_fd, E_NOT_IMPLEMENTED);
    break;
  case REQUEST_REMATCH:
    GS_send_error(client_fd, E_NOT_IMPLEMENTED);
    break;
  case LEAVE_MATCH:
    GS_send_error(client_fd, E_NOT_IMPLEMENTED);
    break;
  case UNSUPPORTED_MESSAGE_TYPE:
  default:
    GS_send_error(client_fd, E_UNSUPPORTED_MESSAGE_TYPE);
  }

  cJSON_Delete(json_request);
}

void GS_handle_create_match(GameServer *gs, int client_fd, cJSON *json_request) {
  Match *match = NULL;
  cJSON *response_json = NULL, *rounds_json = NULL, *mode_json = NULL, *word_len_json = NULL;
  size_t rounds, word_len;
  char *mode_str;
  GameMode game_mode;

  printf("[GS_handle_create_match] json_request: %s\n", cJSON_PrintUnformatted(json_request));
  // parse rounds
  rounds_json = cJSON_GetObjectItem(json_request, "rounds");
  if (rounds_json == NULL) {
    GS_send_error(client_fd, E_MISSING_FIELD("rounds"));
    return;
  }

  if (!cJSON_IsNumber(rounds_json)) {
    GS_send_error(client_fd, E_INVALID_TYPE("rounds", NUMBER));
    return;
  }

  rounds = rounds_json->valueint;

  if (rounds < 1 || rounds > MAX_ROUNDS) {
    GS_send_error(client_fd, E_INVALID_ROUNDS);
    return;
  }

  // parse mode
  mode_json = cJSON_GetObjectItem(json_request, "mode");
  if (mode_json == NULL) {
    GS_send_error(client_fd, E_MISSING_FIELD("mode"));
    return;
  }

  if (!cJSON_IsString(mode_json)) {
    GS_send_error(client_fd, E_INVALID_TYPE("mode", STRING));
    return;
  }

  mode_str = cJSON_GetStringValue(mode_json);
  if (!strcmp("SINGLE", mode_str)) { // TODO: Add support for other modes
    game_mode = SINGLE;
  } else {
    GS_send_error(client_fd, E_UNSUPPORTED_MODE);
    return;
  }

  // parse wordLength
  word_len_json = cJSON_GetObjectItem(json_request, "wordLength");
  if (word_len_json == NULL) {
    GS_send_error(client_fd, E_MISSING_FIELD("wordLength"));
    return;
  }

  if (!cJSON_IsNumber(word_len_json)) {
    GS_send_error(client_fd, E_INVALID_TYPE("wordLength", NUMBER));
    return;
  }

  word_len = word_len_json->valueint;

  if (word_len < MIN_WORD_LEN || word_len > MAX_WORD_LEN) {
    GS_send_error(client_fd, E_INVALID_WORD_LEN);
    return;
  }

  // create match
  match = malloc(sizeof(Match));
  if (match == NULL) {
    perror("malloc");
    exit(1);
  }

  match = new_match(rounds, game_mode, word_len);
  if (match == NULL) {
    printf("[GS_handle_create_match] error: new_match() returned NULL\n");
    return;
  }

  if (gs->head == NULL) {
    gs->head = match;
  } else {
    match->next = gs->head;
    gs->head = match;
  }

  Match_add_player(match, client_fd); // implicitly starts the match

  response_json = cJSON_CreateObject();
  if (response_json == NULL) {
    printf("[GS_handle_create_match] cJSON_CreateObject() failed\n");
    return;
  }
  cJSON_AddStringToObject(response_json, "matchId", match->id);

  GS_send_json(client_fd, response_json);
}

// INFO: Might not need it
Match *GS_get_match_by_player_fd(GameServer *gs, int player_fd) {
  Match *match = gs->head;
  while (match != NULL) {

    // single and multiplayer
    if (match->player1->fd == player_fd) {
      return match;
    }

    // multiplayer
    if (match->player2 != NULL && match->player2->fd == player_fd)
      return match;

    match = match->next;
  }

  return NULL;
}

void GS_send_json(int client_fd, cJSON *json) {
  char *message = cJSON_PrintUnformatted(json);
  cJSON_Delete(json);

  if (message == NULL) {
    printf("[GS_send_json] cJSON_PrintUnformatted() failed");
    return;
  }

  if (send(client_fd, message, strlen(message), 0) == -1) {
    perror("send");
  }

  cJSON_free(message);
}

void GS_send_error(int client_fd, char *reason) {
  cJSON *err_json = json_error(reason);
  GS_send_json(client_fd, err_json);
}

void GS_destroy(GameServer *gs) {
  Match *match = gs->head, *next = NULL;
  while (match != NULL) {
    next = match->next;
    delete_match(match);
    match = next;
  }
  free(gs);
}
