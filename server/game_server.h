#ifndef GAME_SERVER_H
#define GAME_SERVER_H

#include "client.h"
#include "game_logic.h"
#include "game_types.h"
#include "hash_table.h"
#include "room.h"
#include "timer.h"
#include <cjson/cJSON.h>
#include <stddef.h>

#define MAX_ROUNDS 5
#define MAX_CLIENTS 100

// don't ask...
#define STR_HELPER(x) #x
#define STR(x) STR_HELPER(x)

#define STRING "string"
#define NUMBER "number"

#define E_MALFORMED_MESSAGE "Malformed message received"
#define E_UNSUPPORTED_MESSAGE_TYPE "Unsupported message type"
#define E_NOT_IMPLEMENTED "Not Implemented"
#define E_ALREADY_IN_MATCH "Player already in another match"
#define E_MISSING_FIELD(field) "Missing '" field "' field"
#define E_INVALID_STATE(reason) "Invalid state: " reason
#define E_INVALID_TYPE(field, expected_type) "Invalid type of '" field "' field, expected: " expected_type
#define E_INVALID_VALUE(field, reason) "Invalid value received for field '" field "': " reason
#define E_INVALID_ROUNDS "Round number must be between 1 and " STR(MAX_ROUNDS)
#define E_INVALID_WORD_LEN "wordLength must be between " STR(MIN_WORD_LEN) " and " STR(MAX_WORD_LEN)
#define E_NOT_ON_TURN "Opponent is currently on turn"
#define E_REPEATED_GUESS "Repeating guesses is not allowed"
#define E_UNSUPPORTED_MODE "Unsupported mode"
#define E_UNSUPPORTED_FORMAT "Unsupported format"
#define E_PLAYER_NOT_IN_MATCH "Player is not in an active match"
#define E_PLAYER_NOT_IN_ROOM "Player is not in a room"
#define E_ROOM_FULL "Room is full"
#define E_ROOM_NOT_FOUND "Room could not be found"
#define E_ROOM_EMPTY_ON_JOIN "Room has no players and is being deleted"
#define E_UNKNOWN "An unknown error has occured"

typedef enum {
  MALFORMED_MESSAGE = -1,
  UNSUPPORTED_MESSAGE_TYPE,

  // Client
  CREATE_MATCH,
  JOIN_ROOM,
  MAKE_GUESS,
  REQUEST_REMATCH,
  DENY_REMATCH,
  LEAVE_MATCH,
  TYPING,

  // Server
  ROOM_CREATED,
  ROOM_JOINED,
  ROOM_JOIN_FAILED,
  WAIT_OPPONENT_JOIN,
  OPPONENT_DENIED_REMATCH,
  OPPONENT_LEFT,
  MATCH_STARTED,
  ROUND_STARTED,
  WAIT_GUESS,
  WAIT_OPPONENT_GUESS,
  GUESS_RESULT,
  ROUND_FINISHED,
  MATCH_FINISHED,
  OPPONENT_TYPING,
  ERROR,
} MessageType;

typedef struct {
  HashTable *matches;
  HashTable *clients;
  HashTable *rooms;
  Timer *timer_list;
  struct {
    WordStore *five;
    WordStore *six;
    WordStore *seven;
  } word_store;
} GameServer;

GameServer *GS_create(void);
void GS_destroy(GameServer *gs);

void GS_handle_request(GameServer *gs, Client *client);
void GS_handle_create_match(GameServer *gs, Client *client, cJSON *json_request);
void GS_handle_make_guess(GameServer *gs, Client *client, cJSON *json_request);
void GS_handle_join_room(GameServer *gs, Client *client, cJSON *json_request);
void GS_handle_request_rematch(GameServer *gs, Client *client);
void GS_handle_deny_rematch(GameServer *gs, Client *client);
void GS_handle_typing(Client *client, cJSON *json_request);
void GS_handle_leave_match(GameServer *gs, Client *client);

void GS_create_room(GameServer *gs, Match *match, Client *client);
void GS_start_match(GameServer *gs, Match *match, bool is_rematch);
void GS_start_round(GameServer *gs, Match *match);
void GS_end_match(Match *match, Player *disconnected_player);
void GS_end_round(GameServer *gs, Match *match);
bool GS_add_player_to_match(Match *match, Player *player);

void GS_cleanup_after_client_disconnect(GameServer *gs, Client *client);
void GS_cleanup_room(GameServer *gs, Room *room, Player *disconnected_player);
void GS_cleanup_match(GameServer *gs, Match *match);

Player *get_opponent(Player *player1, Player *player2, Player *current);

#endif // !GAME_SERVER_H
