#include "game_server.h"
#include "client.h"
#include "game_logic.h"
#include "game_types.h"
#include "hash_table.h"
#include "json_messages.h"
#include "room.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>

static MessageType parse_message(char *data, size_t size, cJSON **out);
static void start_match(Match *match);
static void start_round(Match *match);
static void end_round(Match *match);
static void add_player_to_match(Match *match, Player *player);

GameServer *GS_create(void) {
  GameServer *gs;

  gs = malloc(sizeof(GameServer));
  if (!gs) {
    perror("malloc");
    exit(1);
  }

  gs->clients = HT_create();
  gs->rooms = HT_create();

  return gs;
}

void GS_destroy(GameServer *gs) {
  HT_destroy(gs->rooms);
  HT_destroy(gs->clients);
  free(gs);
}

void GS_handle_request(GameServer *gs, Client *client) {
  cJSON *json_request = NULL;
  MessageType mt;
  char *data = client->buf_start;
  size_t size = client->payload_size;

  mt = parse_message(data, size, &json_request);

  switch (mt) {
  case MALFORMED_MESSAGE:
    printf("[GS_handle_request] Malformed request received: \"%s\"\n", data);
    if (json_request) {
      cJSON_Delete(json_request);
    }
    send_error(client->fd, E_MALFORMED_MESSAGE);
    return;
  case CREATE_MATCH:
    GS_handle_create_match(client, json_request);
    break;
  case CREATE_ROOM:
    GS_handle_create_room(gs, client);
    break;
  case JOIN_ROOM:
    send_error(client->fd, E_NOT_IMPLEMENTED);
    break;
  case MAKE_GUESS:
    GS_handle_make_guess(client, json_request);
    break;
  case REQUEST_REMATCH:
    send_error(client->fd, E_NOT_IMPLEMENTED);
    break;
  case LEAVE_MATCH:
    send_error(client->fd, E_NOT_IMPLEMENTED);
    break;
  case UNSUPPORTED_MESSAGE_TYPE:
  default:
    send_error(client->fd, E_UNSUPPORTED_MESSAGE_TYPE);
  }

  cJSON_Delete(json_request);
}

void GS_handle_create_match(Client *client, cJSON *json_request) {
  Match *match = NULL;
  cJSON *rounds_json = NULL, *mode_json = NULL, *word_len_json = NULL;
  size_t rounds, word_len;
  char *mode_str;
  GameMode game_mode;

  printf("[GS_handle_create_match] json_request: %s\n", cJSON_PrintUnformatted(json_request));

  if (client->player != NULL || client->player->match != NULL) {
    send_error(client->fd, E_ALREADY_IN_MATCH);
    return;
  }

  // parse rounds
  rounds_json = cJSON_GetObjectItem(json_request, "rounds");
  if (rounds_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("rounds"));
    return;
  }

  if (!cJSON_IsNumber(rounds_json)) {
    send_error(client->fd, E_INVALID_TYPE("rounds", NUMBER));
    return;
  }

  rounds = rounds_json->valueint;

  if (rounds < 1 || rounds > MAX_ROUNDS) {
    send_error(client->fd, E_INVALID_ROUNDS);
    return;
  }

  // parse mode
  mode_json = cJSON_GetObjectItem(json_request, "mode");
  if (mode_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("mode"));
    return;
  }

  if (!cJSON_IsString(mode_json)) {
    send_error(client->fd, E_INVALID_TYPE("mode", STRING));
    return;
  }

  mode_str = cJSON_GetStringValue(mode_json);
  if (!strcmp("SINGLE", mode_str)) {
    game_mode = SINGLE;
  } else if (!strcmp("MULTI_REMOTE", mode_str)) {
    game_mode = MULTI_REMOTE;
  } else {
    send_error(client->fd, E_UNSUPPORTED_MODE);
    return;
  }

  // parse wordLength
  word_len_json = cJSON_GetObjectItem(json_request, "wordLength");
  if (word_len_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("wordLength"));
    return;
  }

  if (!cJSON_IsNumber(word_len_json)) {
    send_error(client->fd, E_INVALID_TYPE("wordLength", NUMBER));
    return;
  }

  word_len = word_len_json->valueint;

  if (word_len < MIN_WORD_LEN || word_len > MAX_WORD_LEN) {
    send_error(client->fd, E_INVALID_WORD_LEN);
    return;
  }

  match = new_match(rounds, game_mode, word_len);
  if (match == NULL) {
    printf("[GS_handle_create_match] error: new_match() returned NULL\n");
    return;
  }

  add_player_to_match(match, client->player); // implicitly starts the match
}

static MessageType parse_message(char *data, size_t size, cJSON **json_out) {
  cJSON *json_type = NULL;
  char *type;
  MessageType mt;

  *json_out = cJSON_ParseWithLength(data, size);
  if (*json_out == NULL) {
    printf("[parse_message] cJSON failed to parse message\n");
    return MALFORMED_MESSAGE;
  }
  printf("[parse_message] json_out: %s\n", cJSON_PrintUnformatted(*json_out));

  json_type = cJSON_GetObjectItem(*json_out, "type");
  if (json_type == NULL) {
    printf("[parse_message] message missing 'type' field\n");
    return MALFORMED_MESSAGE;
  }

  if (!cJSON_IsString(json_type)) {
    printf("[parse_message] message 'type' field is not a string\n");
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

void GS_handle_create_room(GameServer *gs, Client *client) {
  Room *room = new_room();
  HT_set(gs->rooms, KEY(room->id), room);
  room->player1 = client->player;
  printf("[GS_handle_create_room] room created with id: %s", room->id);

  cJSON *room_created_json = json_room_created(room->id);
  send_json(client->fd, room_created_json);
  cJSON_Delete(room_created_json);
}

static void add_player_to_match(Match *match, Player *player) {
  int can_start = 0;
  switch (match->mode) {
  case MULTI_REMOTE:
    if (match->player2 != NULL) {
      printf("[add_player_to_match] error: trying to add a player to a match that has 2 players\n");
    } else {
      match->player2 = player;
      can_start = 1;
    }
    break;
  case SINGLE:
    if (match->player1 != NULL) {
      printf("[add_player_to_match] error: trying to add second player to a match in SINGLE mode\n");
    } else {
      match->player1 = player;
      can_start = 1;
    }
    break;
  }

  if (can_start) {
    start_match(match);
  }
}

void GS_handle_make_guess(Client *client, cJSON *json_request) {
  Match *match = NULL;
  Round *round;
  Player *player, *opponent;
  cJSON *guess_json, *guess_result_json;
  char *guess;
  bool success;
  LetterFeedback *feedback;

  player = client->player;

  if (player != NULL) {
    match = player->match;
  }

  if (match == NULL) {
    send_error(client->fd, E_PLAYER_NOT_IN_MATCH);
    return;
  }

  round = match->rounds[match->round_idx];
  opponent = match->player1 == player ? match->player2 : match->player1;

  assert(round->wc->attempt_count < round->wc->max_attempts);

  if (player != round->on_turn) {
    send_error(client->fd, E_NOT_ON_TURN);
    return;
  }

  guess_json = cJSON_GetObjectItem(json_request, "guess");
  if (guess_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("guess"));
    return;
  }

  if (!cJSON_IsString(guess_json)) {
    send_error(client->fd, E_INVALID_TYPE("guess", STRING));
    return;
  }

  guess = cJSON_GetStringValue(guess_json);
  if (strlen(guess) != match->word_len) {
    send_error(client->fd, E_INVALID_VALUE("guess", "incorrect word length"));
    return;
  }

  feedback = malloc(sizeof(int) * match->word_len);
  if (feedback == NULL) {
    perror("malloc");
    return;
  }

  round->wc->attempt_count++;
  success = evaluate_guess(guess, round->wc->word, feedback, match->word_len);
  guess_result_json = json_guess_result(success, guess, feedback, match->word_len);

  send_json(player->client_fd, guess_result_json);
  if (opponent) {
    send_json(opponent->client_fd, guess_result_json); // TODO: Send more appropriate message
  }
  free(feedback);

  if (!success && !(round->wc->attempt_count >= round->wc->max_attempts)) {
    if (match->mode == MULTI_REMOTE) {
      round->on_turn = opponent;
    }
    return;
  }

  if (success) {
    round->outcome = player == match->player1 ? PLAYER1_WINS : PLAYER2_WINS;
  } else if (round->wc->attempt_count >= round->wc->max_attempts) {
    round->outcome = TIE;
  }
  end_round(match);
}

void GS_end_match(Match *match) {
  char *winner_name;
  switch (match->outcome) {
  case TIE:
    winner_name = "TIE";
    break;
  case PLAYER1_WINS:
    winner_name = match->player1->name;
    break;
  case PLAYER2_WINS:
    winner_name = match->player2->name;
    break;
  }

  cJSON *match_finished_json = json_match_finished(winner_name);

  printf("[GS_end_match] Ending match: (%s)...\n", match->id);

  assert(match->player1 != NULL);
  switch (match->mode) {
  case MULTI_REMOTE:
    send_json(match->player2->client_fd, match_finished_json);
  case SINGLE:
    send_json(match->player1->client_fd, match_finished_json);
  }
  cJSON_Delete(match_finished_json);
}

static void end_round(Match *match) {
  Round *round = match->rounds[match->round_idx];
  cJSON *round_finished_json = json_round_finished(round->outcome, round->wc->word);

  printf("[end_round] Ending round...\n");
  switch (match->mode) {
  case MULTI_REMOTE:
    send_json(match->player2->client_fd, round_finished_json);
  case SINGLE:
    send_json(match->player1->client_fd, round_finished_json);
  }

  cJSON_Delete(round_finished_json);

  if (match->round_idx + 1 < (int)match->round_capacity) {
    start_round(match);
  } else {
    GS_end_match(match);
  }
}

static void start_match(Match *match) {
  cJSON *match_started_json;
  match_started_json = json_match_started(match->id, match->round_capacity, match->word_len);

  printf("[start_match] Starting new match...\n");

  switch (match->mode) {
  case MULTI_REMOTE:
    assert(match->player2 != NULL);
    send_json(match->player2->client_fd, match_started_json);
  case SINGLE:
    assert(match->player1 != NULL);
    send_json(match->player2->client_fd, match_started_json);
  }

  cJSON_Delete(match_started_json);
  start_round(match);
}

static void start_round(Match *match) {
  cJSON *round_started_json = NULL;
  printf("[start_round] Starting new round...\n");

  size_t max_attempts = match->word_len + 1; // TODO: allow for flexible max_attempts
  WordChallenge *wc = new_word_challenge(match->word_len, max_attempts);
  if (wc == NULL) {
    printf("[start_match] error: new_word_challenge() returned NULL\n");
    return;
  }

  Round *round = new_round(wc, match->player1);
  if (round == NULL) {
    printf("[start_match] error: new_round() returned NULL\n");
    return;
  }

  match->rounds[++(match->round_idx)] = round;
  round_started_json = json_round_started(match->round_idx, round->wc->max_attempts);

  switch (match->mode) {
  case MULTI_REMOTE:
    assert(match->player2 != NULL);
    assert(match->player1 != NULL);
    send_json(match->player2->client_fd, round_started_json);
    send_json(match->player1->client_fd, round_started_json);

    Player *not_on_turn;
    if (match->player1 == round->on_turn) {
      not_on_turn = match->player2;
    } else {
      not_on_turn = match->player1;
    }

    send_only_type(round->on_turn->client_fd, STR(WAIT_GUESS));
    send_only_type(not_on_turn->client_fd, STR(WAIT_OPPONENT_GUESS));

    break;

  case SINGLE:
    assert(match->player1 != NULL);
    send_json(match->player1->client_fd, round_started_json);
    send_only_type(match->player1->client_fd, STR(WAIT_GUESS));
    break;
  }
  cJSON_Delete(round_started_json);
}
