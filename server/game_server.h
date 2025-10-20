#ifndef GAME_SERVER_H
#define GAME_SERVER_H

#include "game_types.h"

#define MAX_MATCHES 10
#define MAX_ROUNDS 5

// don't ask...
#define STR_HELPER(x) #x
#define STR(x) STR_HELPER(x)

#define E_INVALID_ROUNDS "Round number must be between 1 and " STR(MAX_ROUNDS) "\n"

typedef enum { CREATE_MATCH, CREATE_ROUND, MAKE_GUESS, UNSUPPORTED_REQUEST } RequestType;

typedef struct {
  Match *matches[MAX_MATCHES];
  int nmatches;
} GameServer;

GameServer *GS_create();
void GS_request(GameServer *gs, int client_fd, char *data, int size);
long GS_create_match(GameServer *gs, int client_fd, int nrounds, char *err);
Match *GS_get_match_by_player(GameServer *gs, int player_fd);
void GS_destroy(GameServer *gs);

#endif // !GAME_SERVER_H
