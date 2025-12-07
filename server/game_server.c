#include "game_server.h"
#include "client.h"
#include "game_logic.h"
#include "json_messages.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>

GameServer *GS_create(void) {
  GameServer *gs;

  gs = malloc(sizeof(GameServer));
  if (!gs) {
    perror("malloc");
    exit(1);
  }

  gs->match_head = NULL;
  gs->clients = calloc(MAX_CLIENTS, sizeof(Client *));

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
    printf("[GS_handle_request] Malformed request received: \"%s\"\n", data);
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
    GS_handle_make_guess(gs, client_fd, json_request);
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
  cJSON *rounds_json = NULL, *mode_json = NULL, *word_len_json = NULL;
  size_t rounds, word_len;
  char *mode_str;
  GameMode game_mode;

  printf("[GS_handle_create_match] json_request: %s\n", cJSON_PrintUnformatted(json_request));

  if (GS_get_match_by_client_fd(gs, client_fd) != NULL) {
    GS_send_error(client_fd, E_ALREADY_IN_MATCH);
    return;
  }

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

  match = new_match(rounds, game_mode, word_len);
  if (match == NULL) {
    printf("[GS_handle_create_match] error: new_match() returned NULL\n");
    return;
  }

  if (gs->match_head == NULL) {
    printf("[GS_handle_create_match] No existing matches, setting gs->match_head to (%p)\n", (void *)match);
    gs->match_head = match;
  } else {
    printf("[GS_handle_create_match] Chaining matches...\n");
    match->next = gs->match_head;
    gs->match_head = match;
  }

  // TODO: make GS_add_player_to_match return an indicator whether we can start the match,
  // and start the match here explicitly.
  GS_add_player_to_match(gs, match, client_fd); // implicitly starts the match
}

void GS_add_player_to_match(GameServer *gs, Match *match, int client_fd) {
  Client *client = gs->clients[client_fd];
  Player *player = new_player(client, match);

  if (player == NULL) {
    printf("[GS_add_player_to_match] error: new_player() returned NULL\n");
    return;
  }

  if (match->player1 == NULL) { // first player
    match->player1 = player;
    if (match->mode == SINGLE) {
      GS_start_match(match);
    }
  } else if (match->player2 == NULL) { // second player
    if (match->mode == SINGLE) {
      printf("[GS_add_player_to_match] error: trying to add second player to a match in SINGLE mode\n");
      return;
    } else { // multiplayer
      match->player2 = player;
      // TODO: start match
    }
  } else {
    printf("[GS_add_player_to_match] error: trying to add a player to a match that has 2 players\n");
    return;
  }
}

void GS_handle_make_guess(GameServer *gs, int client_fd, cJSON *json_request) {
  Match *match = NULL;
  Round *round;
  Player *player, *opponent;
  cJSON *guess_json, *guess_result_json;
  char *guess;
  bool success;
  LetterFeedback *feedback;

  match = GS_get_match_by_client_fd(gs, client_fd);

  if (match == NULL) {
    GS_send_error(client_fd, E_PLAYER_NOT_IN_MATCH);
    return;
  }

  round = match->rounds[match->round_idx];
  player = match->player1->client->fd == client_fd ? match->player1 : match->player2;
  opponent = match->player1 == player ? match->player2 : match->player1;

  assert(round->wc->attempt_count < round->wc->max_attempts);

  if (player != round->on_turn) {
    GS_send_error(client_fd, E_NOT_ON_TURN);
    return;
  }

  guess_json = cJSON_GetObjectItem(json_request, "guess");
  if (guess_json == NULL) {
    GS_send_error(client_fd, E_MISSING_FIELD("guess"));
    return;
  }

  if (!cJSON_IsString(guess_json)) {
    GS_send_error(client_fd, E_INVALID_TYPE("guess", STRING));
    return;
  }

  guess = cJSON_GetStringValue(guess_json);
  if (strlen(guess) != match->word_len) {
    GS_send_error(client_fd, E_INVALID_VALUE("guess", "incorrect word length"));
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

  GS_send_json(player->client->fd, guess_result_json);
  if (opponent) {
    GS_send_json(opponent->client->fd, guess_result_json); // TODO: Send more appropriate message
  }
  free(feedback);

  if (!success && !(round->wc->attempt_count >= round->wc->max_attempts)) {
    if (match->mode == MULTI_REMOTE) { // TODO: Maybe also for MULTI_LOCAL?
      round->on_turn = opponent;
    }
    return;
  }

  if (success) {
    round->outcome = player == match->player1 ? PLAYER1_WINS : PLAYER2_WINS;
  } else if (round->wc->attempt_count >= round->wc->max_attempts) {
    round->outcome = TIE;
  }
  GS_end_round(gs, match, player, opponent);
}

void GS_end_round(GameServer *gs, Match *match, Player *player, Player *opponent) {
  Round *round = match->rounds[match->round_idx];
  cJSON *round_finished_json = json_round_finished(round->outcome, round->wc->word);

  printf("[GS_end_round] Ending round...\n");

  GS_send_json(player->client->fd, round_finished_json);
  if (opponent) {
    GS_send_json(player->client->fd, round_finished_json);
  }
  cJSON_Delete(round_finished_json);

  if (match->round_idx + 1 < (int)match->round_capacity) {
    GS_start_round(match);
  } else {
    GS_end_match(gs, match);
  }
}

void GS_end_match(GameServer *gs, Match *match) {
  Match *current = NULL;
  cJSON *match_finished_json = json_match_finished("unknown"); // TODO: resolve the winner

  printf("[GS_end_match] Ending match: (%s)...\n", match->id);

  assert(match->player1 != NULL);
  GS_send_json(match->player1->client->fd, match_finished_json);
  if (match->player2) {
    GS_send_json(match->player2->client->fd, match_finished_json);
  }
  cJSON_Delete(match_finished_json);

  /* Delete match */
  // 1 match
  if (gs->match_head == match) {
    printf("[GS_end_match] Only one match left, deleting gs->match_head: (%p)\n", (void *)gs->match_head);
    gs->match_head = NULL;
    printf("[GS_end_match] New gs->head: (%p)\n", (void *)gs->match_head);
    delete_match(match);
    return;
  }

  // More than 1 match
  current = gs->match_head;
  while (current) {
    if (current->next == match) {
      current->next = match->next;
      break;
    }
    current = current->next;
  }
  delete_match(match);
}

void GS_start_match(Match *match) {
  cJSON *match_started_json;
  match_started_json = json_match_started(match->id, match->round_capacity, match->word_len);

  printf("[GS_start_match] Starting new match...\n");

  if (match->player1 != NULL) {
    GS_send_json(match->player1->client->fd, match_started_json);
  }
  if (match->player2 != NULL) {
    GS_send_json(match->player2->client->fd, match_started_json);
  }
  cJSON_Delete(match_started_json);
  GS_start_round(match);
}

void GS_start_round(Match *match) {
  cJSON *round_started_json = NULL;
  printf("[GS_start_round] Starting new round...\n");

  size_t max_attempts = match->word_len + 1; // TODO: allow for flexible max_attempts
  WordChallenge *wc = new_word_challenge(match->word_len, max_attempts);
  if (wc == NULL) {
    printf("[Match_start_match] error: new_word_challenge() returned NULL\n");
    return;
  }

  Round *round = new_round(wc, match->player1);
  if (round == NULL) {
    printf("[Match_start_match] error: new_round() returned NULL\n");
    return;
  }

  match->rounds[++(match->round_idx)] = round;
  round_started_json = json_round_started(match->round_idx, round->wc->max_attempts);

  assert(match->player1 != NULL);
  GS_send_json(match->player1->client->fd, round_started_json);
  if (match->player1 == round->on_turn) {
    GS_send_only_type(match->player1->client->fd, STR(WAIT_GUESS));
    if (match->player2 != NULL) {
      GS_send_only_type(match->player1->client->fd, STR(WAIT_OPPONENT_GUESS));
    }
  } else { // player2 is starting
    assert(match->player2 != NULL);
    GS_send_only_type(match->player1->client->fd, STR(WAIT_OPPONENT_GUESS));
    GS_send_only_type(match->player2->client->fd, STR(WAIT_GUESS));
  }
  cJSON_Delete(round_started_json);
}

void delete_match(Match *match) {
  free(match->id);
  if (match->player1 != NULL)
    delete_player(match->player1);
  if (match->player2 != NULL)
    delete_player(match->player2);

  for (int i = 0; i <= match->round_idx; i++) {
    printf("Deleting round idx: %d\n", i);
    delete_round(match->rounds[i]);
  }
  free(match->rounds);
  free(match);
}

Match *GS_get_match_by_client_fd(GameServer *gs, int client_fd) {
  Match *match = gs->match_head;

  printf("[GS_get_match_by_player_fd] getting match by player_fd: %d\n", client_fd);
  while (match != NULL) {

    if (match->player1->client->fd == client_fd) {
      return match;
    }

    // MULTI_REMOTE
    if (match->player2 != NULL && match->player2->client->fd == client_fd)
      return match;

    match = match->next;
  }

  printf("[GS_get_match_by_player_fd] player is not in a match\n");
  return NULL;
}

void GS_send_json(int client_fd, cJSON *json) {
  size_t length;
  uint32_t nlength;
  char *message = cJSON_PrintUnformatted(json);

  if (message == NULL) {
    printf("[GS_send_json] cJSON_PrintUnformatted() failed");
    return;
  }

  // TCP segment length prefix
  length = strlen(message);
  nlength = htonl(length);
  printf("[GS_send_json] sending %zu bytes of data\n", length);
  if (send(client_fd, &nlength, 4, 0) == -1) {
    perror("send");
  }

  // Data
  if (send(client_fd, message, strlen(message), 0) == -1) {
    perror("send");
  }

  cJSON_free(message);
}

void GS_send_only_type(int client_fd, const char *type) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", type);
  GS_send_json(client_fd, json);
  cJSON_Delete(json);
}

void GS_send_error(int client_fd, const char *reason) {
  cJSON *err_json = json_error(reason);
  GS_send_json(client_fd, err_json);
  cJSON_Delete(err_json);
}

void GS_destroy(GameServer *gs) {
  Match *match = gs->match_head, *next = NULL;
  while (match != NULL) {
    next = match->next;
    delete_match(match);
    match = next;
  }

  for (int i = 0; i < MAX_CLIENTS; i++) {
    if (gs->clients[i] == NULL)
      continue;
    delete_client(gs->clients[i]);
  }
  free(gs);
}
