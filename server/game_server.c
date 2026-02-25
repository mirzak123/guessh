#include "game_server.h"
#include "client.h"
#include "game_logic.h"
#include "game_types.h"
#include "hash_table.h"
#include "json_messages.h"
#include "room.h"
#include <_string.h>
#include <assert.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>

static MessageType parse_message(char *data, size_t size, cJSON **out);
static void start_match(Match *match);
static void start_round(Match *match);
static void add_player_to_match(Match *match, Player *player);
static void create_room(GameServer *gs, Match *match, Client *client);
static Outcome calculate_match_outcome(Match *match);

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
    GS_handle_create_match(gs, client, json_request);
    break;
  case JOIN_ROOM:
    GS_handle_join_room(gs, client, json_request);
    break;
  case MAKE_GUESS:
    GS_handle_make_guess(gs, client, json_request);
    break;
  case REQUEST_REMATCH:
    send_error(client->fd, E_NOT_IMPLEMENTED);
    break;
  case LEAVE_MATCH:
    GS_handle_leave_match(gs, client);
    break;
  case TYPING:
    GS_handle_typing(client, json_request);
    break;
  case UNSUPPORTED_MESSAGE_TYPE:
  default:
    send_error(client->fd, E_UNSUPPORTED_MESSAGE_TYPE);
  }

  cJSON_Delete(json_request);
}

void GS_handle_create_match(GameServer *gs, Client *client, cJSON *json_request) {
  Match *match = NULL;
  cJSON *rounds_json = NULL, *mode_json = NULL, *word_len_json = NULL, *player_name_json = NULL;
  size_t rounds, word_len;
  char *mode_str, *player_name_str = NULL;
  GameMode game_mode;

  printf("[GS_handle_create_match] json_request: %s\n", cJSON_PrintUnformatted(json_request));

  if (client->player != NULL && client->player->match != NULL) {
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

  // parse playerName
  if (game_mode == MULTI_REMOTE) {
    player_name_json = cJSON_GetObjectItem(json_request, "playerName");
    if (player_name_json == NULL) {
      send_error(client->fd, E_MISSING_FIELD("playerName"));
      return;
    }

    if (!cJSON_IsString(player_name_json)) {
      send_error(client->fd, E_INVALID_TYPE("playerName", STRING));
      return;
    }

    player_name_str = cJSON_GetStringValue(player_name_json);
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

  Player *player = new_player(client->fd, player_name_str);
  client->player = player;

  if (match->mode == MULTI_REMOTE) {
    create_room(gs, match, client);
  }
  add_player_to_match(match, player); // implicitly starts the match
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
  // printf("[parse_message] json_out: %s\n", cJSON_PrintUnformatted(*json_out));

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
  } else if (!strcmp("TYPING", type)) {
    mt = TYPING;
  } else if (!strcmp("OPPONENT_TYPING", type)) {
    mt = OPPONENT_TYPING;
  } else {
    mt = UNSUPPORTED_MESSAGE_TYPE;
  }

  return mt;
}

void create_room(GameServer *gs, Match *match, Client *client) {
  Room *room = new_room();
  printf("[create_room] room created with id: %s\n", room->id);
  HT_set(gs->rooms, KEY(room->id), room);

  room->match = match;
  room->player1 = client->player;
  match->room_id = strdup(room->id);

  cJSON *room_created_json = json_room_created(room->id);
  send_json(client->fd, room_created_json);
  cJSON_Delete(room_created_json);
}

void GS_handle_join_room(GameServer *gs, Client *client, cJSON *json_request) {
  Room *room;
  char *room_id, *player_name;
  cJSON *room_id_json = NULL, *player_name_json = NULL;

  room_id_json = cJSON_GetObjectItem(json_request, "roomId");
  if (room_id_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("roomId"));
    return;
  }

  if (!cJSON_IsString(room_id_json)) {
    send_error(client->fd, E_INVALID_TYPE("roomId", STRING));
    return;
  }
  room_id = cJSON_GetStringValue(room_id_json);

  player_name_json = cJSON_GetObjectItem(json_request, "playerName");
  if (player_name_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("playerName"));
    return;
  }

  if (!cJSON_IsString(player_name_json)) {
    send_error(client->fd, E_INVALID_TYPE("playerName", STRING));
    return;
  }
  player_name = cJSON_GetStringValue(player_name_json);

  room = (Room *)HT_get(gs->rooms, KEY(room_id));

  if (room == NULL) {
    cJSON *room_join_failed_json = json_room_join_failed(room_id, E_ROOM_NOT_FOUND);
    send_json(client->fd, room_join_failed_json);
    cJSON_Delete(room_join_failed_json);
    return;
  }

  assert(room->player1 != NULL);
  if (room->player2 != NULL) {
    cJSON *room_join_failed_json = json_room_join_failed(room_id, E_ROOM_FULL);
    send_json(client->fd, room_join_failed_json);
    cJSON_Delete(room_join_failed_json);
    return;
  }

  Player *player = new_player(client->fd, player_name);
  client->player = player;
  room->player2 = player;

  cJSON *room_joined_json = json_room_joined(room_id);
  send_json(client->fd, room_joined_json);
  cJSON_Delete(room_joined_json);

  add_player_to_match(room->match, player);
}

void GS_handle_typing(Client *client, cJSON *json_request) {
  Match *match = NULL;

  if (client->player != NULL && client->player->match != NULL) {
    match = client->player->match;
  }

  if (match == NULL) {
    send_error(client->fd, E_PLAYER_NOT_IN_MATCH);
    return;
  }

  Player *opponent = match->player1 == client->player ? match->player2 : match->player1;
  cJSON *value_json = cJSON_GetObjectItem(json_request, "value");

  if (value_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("value"));
    return;
  }

  if (!cJSON_IsString(value_json)) {
    send_error(client->fd, E_INVALID_TYPE("value", STRING));
    return;
  }
  char *value = cJSON_GetStringValue(value_json);

  if (match->mode != MULTI_REMOTE)
    return;

  if (opponent == NULL) {
    return;
  }

  cJSON *opponent_typing_json = json_opponent_typing(value);
  send_json(opponent->client_fd, opponent_typing_json);
  cJSON_Delete(opponent_typing_json);
}

static void add_player_to_match(Match *match, Player *player) {
  bool can_start = false;
  switch (match->mode) {
  case MULTI_REMOTE:
    if (match->player1 == NULL) {
      assert(match->player2 == NULL);
      match->player1 = player;
    } else if (match->player2 == NULL) {
      assert(match->player1 != NULL);
      match->player2 = player;

      if (rand() % 2) {
        match->on_turn = match->player1;
      } else {
        match->on_turn = match->player2;
      }

      can_start = true;
    } else {
      printf("[add_player_to_match] error: trying to add a player to a match that has 2 players\n");
      return;
    }
    break;
  case SINGLE:
    if (match->player1 != NULL) {
      printf("[add_player_to_match] error: trying to add second player to a match in SINGLE mode\n");
      return;
    } else {
      match->player1 = player;
      can_start = true;
    }
    break;
  }

  player->match = match;
  if (can_start) {
    start_match(match);
  }
}

void GS_handle_make_guess(GameServer *gs, Client *client, cJSON *json_request) {
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

  if (match->mode == MULTI_REMOTE && player != match->on_turn) {
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

  bool is_round_finished = success || (round->wc->attempt_count >= round->wc->max_attempts);

  switch (match->mode) {
  case MULTI_REMOTE:
    send_json(player->client_fd, guess_result_json);
    send_json(opponent->client_fd, guess_result_json);

    if (!is_round_finished) {
      send_only_type(opponent->client_fd, STR(WAIT_GUESS));
      send_only_type(player->client_fd, STR(WAIT_OPPONENT_GUESS));
    }
    match->on_turn = opponent;
    break;
  case SINGLE:
    send_json(player->client_fd, guess_result_json);
    if (!is_round_finished) {
      send_only_type(player->client_fd, STR(WAIT_GUESS));
    }
    break;
  }
  free(feedback);

  if (!is_round_finished)
    return;

  if (success) {
    round->outcome = player == match->player1 ? OUTCOME_PLAYER1 : OUTCOME_PLAYER2;
  } else if (round->wc->attempt_count >= round->wc->max_attempts) {
    round->outcome = OUTCOME_NONE;
  }
  GS_end_round(gs, match);
}

void GS_handle_leave_match(GameServer *gs, Client *client) {
  Match *match = NULL;
  if (client->player == NULL || client->player->match == NULL) {
    send_error(client->fd, E_PLAYER_NOT_IN_MATCH);
    return;
  }

  GS_end_match(gs, match, client->player);
}

void GS_end_match(GameServer *gs, Match *match, Player *disconnected_player) {
  cJSON *match_finished_json = NULL;

  switch (match->mode) {
  case MULTI_REMOTE:
    match->outcome = calculate_match_outcome(match);
    break;
  case SINGLE:
    match->outcome = OUTCOME_NONE; // not relevant in SINGLE mode
    break;
  }

  printf("[GS_end_match] Ending match: (%s)...\n", match->id);

  assert(match->player1 != NULL);
  printf("Disconnected player: %d\n", disconnected_player != NULL);
  switch (match->mode) {
  case MULTI_REMOTE:
    if (match->player2 != NULL && match->player2 != disconnected_player) {
      match->player2->match = NULL;
      Outcome outcome; // map outcome to perspective of player 2
      switch (match->outcome) {
      case OUTCOME_PLAYER1:
        outcome = OUTCOME_PLAYER2;
        break;
      case OUTCOME_PLAYER2:
        outcome = OUTCOME_PLAYER1;
        break;
      case OUTCOME_NONE:
        outcome = OUTCOME_NONE;
        break;
      }

      match_finished_json = json_match_finished(outcome, disconnected_player != NULL);
      send_json(match->player2->client_fd, match_finished_json);
      cJSON_Delete(match_finished_json);
    }
    HT_delete(gs->rooms, KEY(match->room_id));

  case SINGLE:
    if (match->player1 != disconnected_player) {
      match->player1->match = NULL;
      match_finished_json = json_match_finished(match->outcome, disconnected_player != NULL);
      send_json(match->player1->client_fd, match_finished_json);
      cJSON_Delete(match_finished_json);
    }
  }
}

void GS_end_round(GameServer *gs, Match *match) {
  Round *round = match->rounds[match->round_idx];
  cJSON *round_finished_json = NULL;

  printf("[end_round] Ending round...\n");
  switch (match->mode) {
    Outcome outcome; // map outcome to perspective of player 2
  case MULTI_REMOTE:
    switch (round->outcome) {
    case OUTCOME_PLAYER1:
      outcome = OUTCOME_PLAYER2;
      break;
    case OUTCOME_PLAYER2:
      outcome = OUTCOME_PLAYER1;
      break;
    case OUTCOME_NONE:
      outcome = OUTCOME_NONE;
      break;
    }
    round_finished_json = json_round_finished(outcome, round->wc->word);
    send_json(match->player2->client_fd, round_finished_json);
    cJSON_Delete(round_finished_json);
  case SINGLE:
    round_finished_json = json_round_finished(round->outcome, round->wc->word);
    send_json(match->player1->client_fd, round_finished_json);
    cJSON_Delete(round_finished_json);
  }

  if (match->round_idx + 1 < (int)match->round_capacity) {
    start_round(match);
  } else {
    GS_end_match(gs, match, NULL);
  }
}

static void start_match(Match *match) {
  cJSON *match_started_json = NULL;

  printf("[start_match] Starting new match...\n");

  switch (match->mode) {
  case MULTI_REMOTE:
    assert(match->player2 != NULL);
    match_started_json = json_match_started(match->id, match->round_capacity, match->word_len, match->player1->name);
    send_json(match->player2->client_fd, match_started_json);
    cJSON_Delete(match_started_json);

    assert(match->player1 != NULL);
    match_started_json = json_match_started(match->id, match->round_capacity, match->word_len, match->player2->name);
    send_json(match->player1->client_fd, match_started_json);
    cJSON_Delete(match_started_json);
    break;

  case SINGLE:
    assert(match->player1 != NULL);
    match_started_json = json_match_started(match->id, match->round_capacity, match->word_len, NULL);
    send_json(match->player1->client_fd, match_started_json);
    cJSON_Delete(match_started_json);
  }

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

  Round *round = new_round(wc);
  if (round == NULL) {
    printf("[start_match] error: new_round() returned NULL\n");
    return;
  }

  match->rounds[++(match->round_idx)] = round;
  round_started_json = json_round_started(match->round_idx + 1, round->wc->max_attempts);

  switch (match->mode) {
  case MULTI_REMOTE:
    assert(match->player2 != NULL);
    assert(match->player1 != NULL);
    send_json(match->player2->client_fd, round_started_json);
    send_json(match->player1->client_fd, round_started_json);

    Player *not_on_turn;
    if (match->player1 == match->on_turn) {
      not_on_turn = match->player2;
    } else {
      not_on_turn = match->player1;
    }

    send_only_type(match->on_turn->client_fd, STR(WAIT_GUESS));
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

static Outcome calculate_match_outcome(Match *match) {
  int outcome = 0;
  for (int i = 0; i <= match->round_idx; i++) {
    Round *round = match->rounds[i];
    if (round->outcome == OUTCOME_PLAYER1)
      outcome++;
    else if (round->outcome == OUTCOME_PLAYER2)
      outcome--;
  }

  if (outcome > 0)
    return OUTCOME_PLAYER1;
  else if (outcome < 0)
    return OUTCOME_PLAYER2;
  return OUTCOME_NONE;
}
