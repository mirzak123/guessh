#include "game_server.h"
#include "game_types.h"
#include "network.h"
#include <_string.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <time.h>

GameServer *GS_create() {
  GameServer *gs;

  gs = malloc(sizeof(GameServer));
  if (!gs) {
    perror("malloc");
    exit(1);
  }

  gs->nmatches = 0;
  return gs;
}

RequestType parse_request_type(char *request, int size) {
  if (!strncmp("CREATE_MATCH", request, 12)) {
    return CREATE_MATCH;
  }
  if (!strncmp("CREATE_ROUND", request, 12)) {
    return CREATE_ROUND;
  }
  if (!strncmp("MAKE_GUESS", request, 10)) {
    return MAKE_GUESS;
  }
  return UNSUPPORTED_REQUEST;
}

void GS_request(GameServer *gs, int client_fd, char *data, int size) {
  RequestType rt = parse_request_type(data, size);

  long status;
  char buf[BUFSIZE];
  sprintf(buf, "Request type: %d\n", rt);

  switch (rt) {
  case CREATE_MATCH:
    if ((status = GS_create_match(gs, client_fd, 1, buf)) != -1)
      sprintf(buf, "Match ID: %ld", status);
    break;

  case UNSUPPORTED_REQUEST:
  default:
    status = -1;
    strlcpy(buf, "error: nsupported request type\n", BUFSIZE);
  }

  if (send(client_fd, buf, strlen(buf), 0) == -1) {
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
