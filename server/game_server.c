#include "game_server.h"
#include "game_types.h"
#include "network.h"
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

  gs->match_number = 0;
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

  if (!strcmp("MATCH_INFO", type)) {
    mt = MATCH_INFO;
  } else if (!strcmp("ROUND_INFO", type)) {
    mt = ROUND_INFO;
  } else if (!strcmp("BYE", type)) {
    mt = BYE;
  } else if (!strcmp("CREATE_ROOM", type)) {
    mt = CREATE_ROOM;
  } else if (!strcmp("CREATE_MATCH", type)) {
    mt = CREATE_MATCH;
  } else if (!strcmp("CREATE_ROUND", type)) {
    mt = CREATE_ROUND;
  } else if (!strcmp("JOIN_ROOM", type)) {
    mt = JOIN_ROOM;
  } else if (!strcmp("MAKE_GUESS", type)) {
    mt = MAKE_GUESS;
  } else if (!strcmp("REQUEST_REMATCH", type)) {
    mt = REQUEST_REMATCH;
  } else if (!strcmp("EXIT_MATCH", type)) {
    mt = EXIT_MATCH;
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
  } else if (!strcmp("ACCEPTED_GUESS", type)) {
    mt = ACCEPTED_GUESS;
  } else if (!strcmp("INVALID_GUESS", type)) {
    mt = INVALID_GUESS;
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
  char response_buf[BUFSIZE];
  MessageType mt;

  mt = GS_parse_message(data, size, &json_request);
  printf("[GS_handle_request] json_request: %s\n", cJSON_PrintUnformatted(json_request));

  switch (mt) {
  case MALFORMED_MESSAGE:
    printf("[GS_handle_request] Malformed request received\n");
    if (json_request) {
      cJSON_Delete(json_request);
    }
    return;
  case CREATE_MATCH:
    GS_handle_create_match(gs, client_fd, json_request);
    break;
  case CREATE_ROOM:
    sprintf(response_buf, "Room created\n");
    break;
  case JOIN_ROOM:
    sprintf(response_buf, "Joining room...\n");
    break;
  case MAKE_GUESS:
    sprintf(response_buf, "Verifying guess...\n");
    break;
  case REQUEST_REMATCH:
    sprintf(response_buf, "Rematch requested\n");
    break;
  case EXIT_MATCH:
    sprintf(response_buf, "Leaving match...\n");
    break;
  case UNSUPPORTED_MESSAGE_TYPE:
  default:
    strlcpy(response_buf, "error: unsupported request type\n", BUFSIZE);
  }

  cJSON_Delete(json_request);
}

void GS_handle_create_match(GameServer *gs, int client_fd, cJSON *json_request) {
  Match *match = NULL;
  cJSON *response_json = NULL, *rounds_json = NULL, *mode_json = NULL;
  int rounds;
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

  // create match
  match = malloc(sizeof(Match));
  if (match == NULL) {
    perror("malloc");
    exit(1);
  }

  match->id = malloc(sizeof(long));
  sprintf(match->id, "%lu", time(NULL));

  match->mode = game_mode;
  match->player = (Player){client_fd};
  match->round_current = 0;
  match->round_capacity = rounds;

  // TODO: add check for match limit
  gs->matches[(gs->match_number)++] = match;

  response_json = cJSON_CreateObject();
  if (response_json == NULL) {
    printf("[GS_handle_create_match] cJSON_CreateObject() failed\n");
    return;
  }
  cJSON_AddStringToObject(response_json, "matchId", match->id);

  GS_send_json(client_fd, response_json);
}

Match *GS_get_match_by_player(GameServer *gs, int player_fd) {
  for (int i = 0; i < gs->match_number; i++) {
    if (gs->matches[i]->player.fd == player_fd) {
      return gs->matches[i];
    }
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
  cJSON *err_json = NULL;

  err_json = cJSON_CreateObject();
  cJSON_AddStringToObject(err_json, "type", "ERROR");
  cJSON_AddStringToObject(err_json, "reason", reason);
  GS_send_json(client_fd, err_json);
}

void GS_destroy(GameServer *gs) {
  Match *match;
  for (int i = 0; i < gs->match_number; i++) {
    match = gs->matches[i];
    free(match->id);
  }
  free(gs);
}
