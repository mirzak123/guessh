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

  gs->nmatches = 0;
  return gs;
}

MessageType GS_parse_message(char *data, size_t size, cJSON *json_out) {
  cJSON *json_type = NULL;
  char *type;
  MessageType mt;

  json_out = cJSON_ParseWithLength(data, size);
  if (json_out == NULL) {
    printf("cJSON failed to parse message\n");
    return -1;
  }

  json_type = cJSON_GetObjectItem(json_out, "type");
  if (json_type == NULL) {
    printf("cJSON cannot get type from message");
    return -1;
  }

  // WARN: might have to check if it's really a string
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

  cJSON_free(json_type);
  return mt;
}

void GS_request(GameServer *gs, int client_fd, char *data, size_t size) {
  cJSON *json_request = NULL;
  long status;
  char response_buf[BUFSIZE];
  MessageType mt;

  mt = GS_parse_message(data, size, json_request);

  switch (mt) {
  case MALFORMED_MESSAGE:
    printf("[GS] Malformed request received");
    if (json_request) {
      cJSON_free(json_request);
    }
    return;
  case CREATE_MATCH:
    if ((status = GS_create_match(gs, client_fd, 1, response_buf)) != -1)
      sprintf(response_buf, "Match ID: %ld\n", status);
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

  cJSON_free(json_request);

  if (send(client_fd, response_buf, strlen(response_buf), 0) == -1) {
    perror("send");
  }
}

long GS_create_match(GameServer *gs, int client_fd, int nrounds, char *err) {
  Match *match;

  if (nrounds < 1 || nrounds > MAX_ROUNDS) {
    strlcpy(err, E_INVALID_ROUNDS, sizeof E_INVALID_ROUNDS);
    return -1;
  }

  match = malloc(sizeof(Match));
  if (match == NULL) {
    perror("malloc");
    exit(1);
  }

  match->id = time(NULL);
  match->mode = SINGLE; // INFO: Only one supported.
  match->player = (Player){client_fd};
  match->round_current = 0;
  match->round_capacity = nrounds;

  // TODO: add check for match limit
  // WARN: gs->nmatches++ possibly increases the pointer, not the value
  gs->matches[gs->nmatches++] = match;

  return match->id;
}

Match *GS_get_match_by_player(GameServer *gs, int player_fd) {
  for (int i = 0; i < gs->nmatches; i++) {
    if (gs->matches[i]->player.fd == player_fd) {
      return gs->matches[i];
    }
  }
  return NULL;
}

void GS_destroy(GameServer *gs) { free(gs); }
