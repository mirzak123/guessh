#include "game_server.h"
#include "client.h"
#include "game_logic.h"
#include "game_types.h"
#include "hash_table.h"
#include "json_messages.h"
#include "room.h"
#include "timer.h"
#include <assert.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <unistd.h>

#define READY_FOR_TURN_TIMEOUT 10

typedef struct {
  GameServer *gs;
  Match *match;
} TurnTimerData;

static MessageType parse_client_event(char *data, size_t size, cJSON **out);
static Outcome calculate_match_outcome(Match *match);
static void calculate_round_points(Round *round);
static WordStore *get_word_store(GameServer *gs, size_t word_len);
static const char **get_round_words(Round *round);
static bool already_guessed(char *word, char **guesses, size_t len);
static bool is_round_finished(Round *round);
static void add_guess_attempt(Round *round, char *guess);

static void swap_turn(Match *match);
static void start_turn(Match *match);
static void send_guess_result(Match *match, char *guess);
static TimerFireAction expire_turn_timer(TurnTimerData *timer_data);
static TimerFireAction expire_post_round_timer(Match *match);

GameServer *GS_create(void) {
  GameServer *gs;

  gs = calloc(1, sizeof(GameServer));
  if (gs == NULL) {
    perror("calloc");
    exit(1);
  }

  gs->matches = HT_create();
  gs->clients = HT_create();
  gs->rooms = HT_create();
  gs->timer_list = calloc(1, sizeof(TimerList));

  gs->word_store.five_secret = new_word_store(FIVE_LETTER_SECRET_WORD_FILE, 5);
  gs->word_store.six_secret = new_word_store(SIX_LETTER_SECRET_WORD_FILE, 6);
  gs->word_store.seven_secret = new_word_store(SEVEN_LETTER_SECRET_WORD_FILE, 7);

  return gs;
}

void GS_destroy(GameServer *gs) {
  HT_destroy(gs->rooms, (ValueDestructor)delete_room);
  HT_destroy(gs->clients, (ValueDestructor)delete_client);
  HT_destroy(gs->matches, (ValueDestructor)delete_match);
  free(gs->timer_list);

  delete_word_store(gs->word_store.five_secret);
  delete_word_store(gs->word_store.six_secret);
  delete_word_store(gs->word_store.seven_secret);

  free(gs);
}

void GS_handle_request(GameServer *gs, Client *client) {
  cJSON *json_request = NULL;
  MessageType mt;
  char *data = client->buf_start;
  size_t size = client->payload_size;

  mt = parse_client_event(data, size, &json_request);

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
    GS_handle_request_rematch(gs, client);
    break;
  case DENY_REMATCH:
    GS_handle_deny_rematch(gs, client);
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
  cJSON *rounds_json = NULL, *mode_json = NULL, *format_json = NULL, *word_len_json = NULL, *player_name_json = NULL,
        *seconds_per_turn_json = NULL;
  size_t rounds, word_len, seconds_per_turn = 0;
  char *mode_str, *format_str, *player_name_str = NULL;
  GameMode game_mode;
  GameFormat game_format;

  char *debug_json = cJSON_PrintUnformatted(json_request);
  if (debug_json) {
    printf("[GS_handle_create_match] json_request: %s\n", debug_json);
    free(debug_json);
  }

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
  } else if (!strcmp("MULTI_LOCAL", mode_str)) {
    game_mode = MULTI_LOCAL;
  } else {
    send_error(client->fd, E_UNSUPPORTED_MODE);
    return;
  }

  // parse format
  format_json = cJSON_GetObjectItem(json_request, "format");
  if (format_json == NULL) {
    send_error(client->fd, E_MISSING_FIELD("format"));
    return;
  }

  if (!cJSON_IsString(format_json)) {
    send_error(client->fd, E_INVALID_TYPE("format", STRING));
    return;
  }

  format_str = cJSON_GetStringValue(format_json);
  if (!strcmp("WORDLE", format_str)) {
    game_format = WORDLE;
  } else if (!strcmp("QUORDLE", format_str)) {
    game_format = QUORDLE;
  } else {
    send_error(client->fd, E_UNSUPPORTED_FORMAT);
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

  // parse secondsPerTurn
  seconds_per_turn_json = cJSON_GetObjectItem(json_request, "secondsPerTurn");
  if (seconds_per_turn_json != NULL) {
    if (!cJSON_IsNumber(seconds_per_turn_json)) {
      send_error(client->fd, E_INVALID_TYPE("secondsPerTurn", NUMBER));
      return;
    }

    seconds_per_turn = seconds_per_turn_json->valueint;

    printf("seconds per turn: %lu\n", seconds_per_turn);
    if (seconds_per_turn < MIN_SECONDS_PER_TURN || seconds_per_turn > MAX_SECONDS_PER_TURN) {
      send_error(client->fd, E_INVALID_SECONDS_PER_TURN);
      return;
    }
  }

  match = new_match(game_mode, game_format, rounds, word_len);
  if (match == NULL) {
    printf("[GS_handle_create_match] error: new_match() returned NULL\n");
    return;
  }
  HT_set(gs->matches, KEY(match->id), match);

  if (seconds_per_turn > 0) {
    TurnTimerData *timer_data = malloc(sizeof(TurnTimerData));
    if (timer_data == NULL) {
      perror("timer_data malloc");
      delete_match(match);
    }
    timer_data->gs = gs;
    timer_data->match = match;
    match->turn_timer = new_timer(gs->timer_list, (TimerCallbackFunc)expire_turn_timer, timer_data, seconds_per_turn);

    if (game_mode == MULTI_REMOTE) {
      match->post_round_timer =
          new_timer(gs->timer_list, (TimerCallbackFunc)expire_post_round_timer, match, READY_FOR_TURN_TIMEOUT);
    }
  }

  if (client->player != NULL) {
    // client was already in another match and had an assigned player
    delete_player(client->player);
    client->player = NULL;
  }

  Player *player = new_player(client->fd, player_name_str);
  if (player == NULL) {
    send_error(client->fd, E_UNKNOWN);
    return;
  }
  client->player = player;

  if (match->mode == MULTI_REMOTE) {
    GS_create_room(gs, match, client);
  }

  bool can_start = GS_add_player_to_match(match, player);
  if (can_start) {
    GS_start_match(gs, match, false);
  }
}

void GS_cleanup_after_client_disconnect(GameServer *gs, Client *client) {
  assert(gs != NULL);
  assert(client != NULL);

  printf("Cleaning up after client [fd: %d]\n", client->fd);

  if (client->player != NULL) {
    Player *player = client->player;

    if (player->match != NULL) {
      GS_end_match(gs, player->match, player);
      GS_cleanup_match(gs, player->match);
    }

    if (player->room != NULL) {
      GS_cleanup_room(gs, player->room, player);
    }
  } else {
    printf("Client [fd: %d] has no player associated. Skipping room and match cleanup\n", client->fd);
  }

  close(client->fd);
  HT_delete(gs->clients, KEY(client->fd));
  delete_client(client);
}

void GS_cleanup_room(GameServer *gs, Room *room, Player *disconnected_player) {
  assert(room != NULL);
  assert(disconnected_player != NULL);
  printf("Cleaning up room [id: %s]\n", room->id);

  Player *opponent = get_opponent(room->player1, room->player2, disconnected_player);
  if (opponent != NULL) {
    send_only_type(opponent->client_fd, STR(OPPONENT_LEFT));
    opponent->room = NULL;
    opponent->waiting_ready_for_turn = false;
  }
  disconnected_player->room = NULL;
  disconnected_player->waiting_ready_for_turn = false;

  HT_delete(gs->rooms, KEY(room->id));
  delete_room(room);
}

void GS_cleanup_match(GameServer *gs, Match *match) {
  assert(match != NULL);
  printf("Cleaning up match [id: %s]\n", match->id);

  if (match->player1 != NULL) {
    match->player1->match = NULL;
  }

  if (match->player2 != NULL) {
    match->player2->match = NULL;
  }

  HT_delete(gs->matches, KEY(match->id));
  delete_match(match);
}

MessageType parse_client_event(char *data, size_t size, cJSON **json_out) {
  cJSON *json_type = NULL;
  char *type;
  MessageType mt;

  *json_out = cJSON_ParseWithLength(data, size);
  if (*json_out == NULL) {
    printf("[%s] cJSON failed to parse message\n", __FUNCTION__);
    return MALFORMED_MESSAGE;
  }

  json_type = cJSON_GetObjectItem(*json_out, "type");
  if (json_type == NULL) {
    printf("[%s] message missing 'type' field\n", __FUNCTION__);
    return MALFORMED_MESSAGE;
  }

  if (!cJSON_IsString(json_type)) {
    printf("[%s] message 'type' field is not a string\n", __FUNCTION__);
    return MALFORMED_MESSAGE;
  }
  type = cJSON_GetStringValue(json_type);

  if (!strcmp("CREATE_MATCH", type)) {
    mt = CREATE_MATCH;
  } else if (!strcmp("JOIN_ROOM", type)) {
    mt = JOIN_ROOM;
  } else if (!strcmp("MAKE_GUESS", type)) {
    mt = MAKE_GUESS;
  } else if (!strcmp("REQUEST_REMATCH", type)) {
    mt = REQUEST_REMATCH;
  } else if (!strcmp("DENY_REMATCH", type)) {
    mt = DENY_REMATCH;
  } else if (!strcmp("LEAVE_MATCH", type)) {
    mt = LEAVE_MATCH;
  } else if (!strcmp("TYPING", type)) {
    mt = TYPING;
  } else {
    mt = UNSUPPORTED_MESSAGE_TYPE;
  }

  return mt;
}

void GS_create_room(GameServer *gs, Match *match, Client *client) {
  Room *room = new_room();
  printf("[create_room] room created with id: %s\n", room->id);
  HT_set(gs->rooms, KEY(room->id), room);

  room->match = match;
  room->player1 = client->player;
  room->player1->room = room;
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

  if (room->player1 == NULL) {
    cJSON *room_join_failed_json = json_room_join_failed(room_id, E_ROOM_EMPTY_ON_JOIN);
    send_json(client->fd, room_join_failed_json);
    cJSON_Delete(room_join_failed_json);

    HT_delete(gs->rooms, KEY(room->id));
    delete_room(room);
    return;
  }

  if (room->player2 != NULL) {
    cJSON *room_join_failed_json = json_room_join_failed(room_id, E_ROOM_FULL);
    send_json(client->fd, room_join_failed_json);
    cJSON_Delete(room_join_failed_json);
    return;
  }

  if (client->player != NULL) {
    delete_player(client->player);
    client->player = NULL;
  }

  Player *player = new_player(client->fd, player_name);
  if (player == NULL) {
    send_error(client->fd, E_UNKNOWN);
    return;
  }

  client->player = player;
  room->player2 = player;
  room->player2->room = room;

  cJSON *room_joined_json = json_room_joined(room_id);
  send_json(client->fd, room_joined_json);
  cJSON_Delete(room_joined_json);

  GS_add_player_to_match(room->match, player);
  GS_start_match(gs, room->match, false);
}

void GS_handle_request_rematch(GameServer *gs, Client *client) {
  assert(client->player != NULL);
  Player *player = client->player;
  Player *opponent;
  Room *room = player->room;
  Match *old_match, *match;

  if (player == NULL) {
    send_error(client->fd, E_PLAYER_NOT_IN_MATCH);
    return;
  }

  old_match = player->match;

  if (old_match == NULL) {
    // silently ignore, rematch likely denied by other player
    return;
  }

  if (old_match->mode == MULTI_REMOTE) {
    if (room == NULL) {
      send_error(client->fd, E_PLAYER_NOT_IN_ROOM);
      return;
    }
    opponent = get_opponent(room->player1, room->player2, player);

    player->wants_rematch = true;
    if (!opponent->wants_rematch) {
      return;
    }
  }

  match = new_match(old_match->mode, old_match->format, old_match->round_capacity, old_match->word_len);
  if (match == NULL) {
    printf("[GS_handle_request_rematch] error: new_match() returned NULL\n");
    return;
  }
  HT_set(gs->matches, KEY(match->id), match);

  GS_add_player_to_match(match, player);
  switch (match->mode) {
  case MULTI_REMOTE:
    GS_add_player_to_match(match, opponent);
    match->remote.match_starter = get_opponent(player, opponent, old_match->remote.match_starter);
    break;
  case MULTI_LOCAL:
    match->local.p1_start_match = !old_match->local.p1_start_match;
    break;
  default:
    break;
  }

  GS_start_match(gs, match, true);

  HT_delete(gs->matches, KEY(old_match->id));
  delete_match(old_match);
}

void GS_handle_deny_rematch(GameServer *gs, Client *client) {
  assert(client->player != NULL);
  Player *player = client->player;
  Player *opponent;
  Match *match = client->player->match;
  Room *room = client->player->room;

  if (match == NULL) {
    // silently ignore, rematch likely denied by other player
    return;
  }

  switch (match->mode) {
  case MULTI_REMOTE:
    player->wants_rematch = false;

    opponent = get_opponent(match->player1, match->player2, player);
    if (opponent) {
      opponent->match = NULL;
      opponent->room = NULL;
      send_only_type(opponent->client_fd, STR(OPPONENT_DENIED_REMATCH));
    }

    if (room != NULL) {
      HT_delete(gs->rooms, KEY(room->id));
      delete_room(room);
    }

    /* fallthrough */
  case MULTI_LOCAL:
  case SINGLE:
    player->match = NULL;
    break;
  }

  HT_delete(gs->matches, KEY(match->id));
  delete_match(match);
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

  Player *opponent = get_opponent(match->player1, match->player2, client->player);
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

bool GS_add_player_to_match(Match *match, Player *player) {
  bool can_start = false;
  switch (match->mode) {
  case MULTI_REMOTE:
    if (match->player1 == NULL) {
      assert(match->player2 == NULL);
      match->player1 = player;
    } else if (match->player2 == NULL) {
      assert(match->player1 != NULL);
      match->player2 = player;
      can_start = true;
    } else {
      printf("[add_player_to_match] error: trying to add a player to a match that has 2 players\n");
      return false;
    }

    break;
  case MULTI_LOCAL:
    if (match->player1 != NULL) {
      printf("[add_player_to_match] error: trying to add second player to a match in MULTI_LOCAL mode\n");
      return false;
    }
    /* fallthrough */
  case SINGLE:
    if (match->player1 != NULL) {
      printf("[add_player_to_match] error: trying to add second player to a match in SINGLE mode\n");
      return false;
    }

    match->player1 = player;
    can_start = true;
    break;
  }

  player->match = match;
  player->wants_rematch = false;

  return can_start;
}

void GS_handle_make_guess(GameServer *gs, Client *client, cJSON *json_request) {
  Match *match = NULL;
  Round *round;
  Player *player, *opponent;
  cJSON *guess_json;
  char *guess;
  bool player1_on_turn;

  player = client->player;

  if (player != NULL) {
    match = player->match;
  }

  if (match == NULL) {
    send_error(client->fd, E_PLAYER_NOT_IN_MATCH);
    return;
  }

  round = match->rounds[match->round_idx];
  opponent = get_opponent(match->player1, match->player2, player);

  printf("attempt_count: %lu\n", round->attempt_count);
  printf("attempt_count: %lu\n", round->max_attempts);
  assert(round->attempt_count < round->max_attempts);

  if (match->mode == MULTI_REMOTE && player != match->remote.on_turn) {
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

  if (already_guessed(guess, round->guess_attempts, round->attempt_count)) {
    send_error(client->fd, E_REPEATED_GUESS);
    if (opponent != NULL) {
      cJSON *opponent_typing_json = json_opponent_typing("");
      send_json(opponent->client_fd, opponent_typing_json);
      cJSON_Delete(opponent_typing_json);
    }
    return;
  }

  switch (match->mode) {
  case SINGLE:
    player1_on_turn = true;
    break;
  case MULTI_LOCAL:
    player1_on_turn = match->local.p1_on_turn;
    break;
  case MULTI_REMOTE:
    player1_on_turn = player == match->player1;
    break;
  }

  add_guess_attempt(round, guess);
  size_t solved_num = evaluate_guess(guess, round->wc_list, round->wc_num, player1_on_turn);
  round->solved_num += solved_num;

  if (solved_num == 0) {
    swap_turn(match);
  }

  send_guess_result(match, guess);

  if (is_round_finished(round)) {
    GS_end_round(gs, match); // TODO: arm the ready for turn timer if match isn't over
  } else {
    start_turn(match);
  }
}

void GS_handle_leave_match(GameServer *gs, Client *client) {
  if (client->player == NULL || client->player->match == NULL) {
    send_error(client->fd, E_PLAYER_NOT_IN_MATCH);
    return;
  }

  Room *room = client->player->room;
  Match *match = client->player->match;

  HT_delete(gs->rooms, KEY(room->id));
  delete_room(room);
  client->player->room = NULL;

  HT_delete(gs->matches, KEY(match->id));
  delete_match(match);
  client->player->match = NULL;
}

void GS_handle_ready_for_turn(GameServer *gs, Client *client) {
  (void)gs;
  Player *player = client->player, *opponent;

  if (client->player == NULL || !client->player->waiting_ready_for_turn) {
    send_error(client->fd, E_NOT_WAITING_FOR_READY_FOR_TURN);
    return;
  }

  Match *match = client->player->match;

  assert(match != NULL);

  switch (match->mode) {
  case MULTI_REMOTE:
    opponent = get_opponent(match->player1, match->player2, player);
    if (opponent->waiting_ready_for_turn) {
      player->waiting_ready_for_turn = false;
      opponent->waiting_ready_for_turn = false;
      start_turn(match);
    }
    break;
  case SINGLE:
  case MULTI_LOCAL:
    break;
  }
}

void GS_end_match(GameServer *gs, Match *match, Player *disconnected_player) {
  (void)gs;

  cJSON *match_finished_json = NULL;

  if (match->turn_timer != NULL) {
    Timer_disarm(match->turn_timer);
  }

  switch (match->mode) {
  case MULTI_REMOTE:
    match->outcome = calculate_match_outcome(match);
    break;
  case MULTI_LOCAL:
    match->outcome = calculate_match_outcome(match);
    break;
  case SINGLE:
    match->outcome = OUTCOME_NONE; // not relevant in SINGLE mode
    break;
  }

  printf("[GS_end_match] Ending match: (%s)...\n", match->id);
  assert(match->player1 != NULL);

  if (match->mode == MULTI_REMOTE) {
    if (match->player2 != NULL && match->player2 != disconnected_player) {
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
  }

  if (match->player1 != disconnected_player) {
    match_finished_json = json_match_finished(match->outcome, disconnected_player != NULL);
    send_json(match->player1->client_fd, match_finished_json);
    cJSON_Delete(match_finished_json);
  }
}

TimerFireAction GS_end_round(GameServer *gs, Match *match) {
  Round *round = match->rounds[match->round_idx];
  cJSON *round_finished_json = NULL;

  calculate_round_points(round);

  const char **words = get_round_words(round);
  if (words == NULL) {
    printf("get_round_words returned NULL\n");
    return false;
  }

  printf("[end_round] Ending round...\n");
  switch (match->mode) {
  case MULTI_REMOTE:
    round_finished_json = json_round_finished(round->points * -1, words, round->wc_num);
    send_json(match->player2->client_fd, round_finished_json);
    cJSON_Delete(round_finished_json);

    round_finished_json = json_round_finished(round->points, words, round->wc_num);
    send_json(match->player1->client_fd, round_finished_json);
    cJSON_Delete(round_finished_json);

    match->remote.round_starter = get_opponent(match->player1, match->player2, match->remote.round_starter);

    break;
  case MULTI_LOCAL:
    match->local.p1_start_round = !match->local.p1_start_round;
    /* fallthrough */
  case SINGLE:
    round_finished_json = json_round_finished(round->points, words, round->wc_num);
    send_json(match->player1->client_fd, round_finished_json);
    cJSON_Delete(round_finished_json);
  }

  free(words);

  if (match->round_idx + 1 < (int)match->round_capacity) {
    GS_start_round(gs, match);
    return true;
  } else {
    GS_end_match(gs, match, NULL);
    return false;
  }
}

void GS_start_match(GameServer *gs, Match *match, bool is_rematch) {
  cJSON *match_started_json = NULL;

  printf("[start_match] Starting new match...\n");

  switch (match->mode) {
  case MULTI_REMOTE:
    assert(match->player2 != NULL);
    match_started_json = json_match_started(match->id, match->format, match->round_capacity, match->word_len, match->turn_timer,
                                            match->player1->name);
    send_json(match->player2->client_fd, match_started_json);
    cJSON_Delete(match_started_json);

    assert(match->player1 != NULL);
    match_started_json = json_match_started(match->id, match->format, match->round_capacity, match->word_len, match->turn_timer,
                                            match->player2->name);
    send_json(match->player1->client_fd, match_started_json);
    cJSON_Delete(match_started_json);

    if (!is_rematch) {
      if (rand() % 2) {
        match->remote.match_starter = match->player1;
      } else {
        match->remote.match_starter = match->player2;
      }
    }
    match->remote.round_starter = match->remote.match_starter;

    break;
  case MULTI_LOCAL:
    if (!is_rematch) {
      match->local.p1_start_match = rand() % 2;
    }
    match->local.p1_start_round = match->local.p1_start_match;
    printf("p1_start_match: %d\n", match->local.p1_start_round);
    /* fallthrough */
  case SINGLE:
    assert(match->player1 != NULL);
    match_started_json =
        json_match_started(match->id, match->format, match->round_capacity, match->word_len, match->turn_timer, NULL);
    send_json(match->player1->client_fd, match_started_json);
    cJSON_Delete(match_started_json);
  }

  GS_start_round(gs, match);
  if (match->turn_timer != NULL) {
    Timer_arm(match->turn_timer);
  }
}

void GS_start_round(GameServer *gs, Match *match) {
  cJSON *round_started_json = NULL;
  printf("[start_round] Starting new round...\n");

  WordStore *store = get_word_store(gs, match->word_len);

  size_t max_attempts = match->word_len;
  size_t wc_num;
  switch (match->format) {
  case WORDLE:
    wc_num = 1;
    max_attempts += 1;
    break;
  case QUORDLE:
    wc_num = 4;
    max_attempts += 4;
    break;
  }

  WordChallenge **wc_list = malloc(sizeof(WordChallenge *) * wc_num);

  for (size_t i = 0; i < wc_num; i++) {
    WordChallenge *wc = new_word_challenge(store);
    if (wc == NULL) {
      printf("[start_match] error: new_word_challenge() returned NULL\n");
      return;
    }
    wc_list[i] = wc;
  }

  Round *round = new_round(wc_list, wc_num, max_attempts);
  if (round == NULL) {
    printf("[start_match] error: new_round() returned NULL\n");
    return;
  }

  match->rounds[++(match->round_idx)] = round;
  round_started_json = json_round_started(match->round_idx + 1, round->max_attempts);

  switch (match->mode) {
  case MULTI_REMOTE:
    assert(match->player2 != NULL);
    assert(match->player1 != NULL);
    send_json(match->player2->client_fd, round_started_json);
    send_json(match->player1->client_fd, round_started_json);

    match->remote.on_turn = match->remote.round_starter;
    break;
  case MULTI_LOCAL:
    assert(match->player1 != NULL);
    send_json(match->player1->client_fd, round_started_json);

    match->local.p1_on_turn = match->local.p1_start_round;
    printf("player1_on_turn: %d\n", match->local.p1_on_turn);

    break;
  case SINGLE:
    assert(match->player1 != NULL);
    send_json(match->player1->client_fd, round_started_json);
    break;
  }
  cJSON_Delete(round_started_json);

  start_turn(match);
}

Player *get_opponent(Player *player1, Player *player2, Player *current) { return player1 == current ? player2 : player1; }

Outcome calculate_match_outcome(Match *match) {
  int total_points = 0;
  for (int i = 0; i <= match->round_idx; i++) {
    total_points += match->rounds[i]->points;
  }

  if (total_points > 0)
    return OUTCOME_PLAYER1;
  else if (total_points < 0)
    return OUTCOME_PLAYER2;
  return OUTCOME_NONE;
}

void calculate_round_points(Round *round) {
  for (size_t i = 0; i < round->wc_num; i++) {
    WordChallenge *wc = round->wc_list[i];
    switch (wc->solved_by) {
    case OUTCOME_NONE:
      break;
    case OUTCOME_PLAYER1:
      round->points++;
      break;
    case OUTCOME_PLAYER2:
      round->points--;
      break;
    }
  }
}

WordStore *get_word_store(GameServer *gs, size_t word_len) {
  switch (word_len) {
  case 5:
    return gs->word_store.five_secret;
  case 6:
    return gs->word_store.six_secret;
  case 7:
    return gs->word_store.seven_secret;
  }
  printf("Tried to get word store for unsupported word length: %lu\n", word_len);
  exit(EXIT_FAILURE);
}

const char **get_round_words(Round *round) {
  const char **words = malloc(sizeof(char *) * round->wc_num);
  if (words == NULL) {
    perror("malloc");
    return NULL;
  }

  for (size_t i = 0; i < round->wc_num; i++) {
    words[i] = round->wc_list[i]->word;
  }

  return words;
}

bool already_guessed(char *word, char **guesses, size_t len) {
  for (size_t i = 0; i < len; i++) {
    if (!strcmp(guesses[i], word)) {
      return true;
    }
  }
  return false;
}

void send_guess_result(Match *match, char *guess) {
  Player *player, *opponent;
  cJSON *guess_result_json = json_guess_result(guess, match->rounds[match->round_idx], match->word_len);

  switch (match->mode) {
  case MULTI_REMOTE:
    player = match->remote.on_turn;
    opponent = get_opponent(match->player1, match->player2, player);
    send_json(player->client_fd, guess_result_json);
    send_json(opponent->client_fd, guess_result_json);
    break;
  case MULTI_LOCAL:
    send_json(match->player1->client_fd, guess_result_json);
    break;
  case SINGLE:
    send_json(match->player1->client_fd, guess_result_json);
    break;
  }

  cJSON_Delete(guess_result_json);
}

void swap_turn(Match *match) {
  Player *player, *opponent;

  switch (match->mode) {
  case MULTI_REMOTE:
    opponent = match->remote.on_turn;
    player = get_opponent(match->player1, match->player2, opponent);
    match->remote.on_turn = player;
    break;
  case MULTI_LOCAL:
    match->local.p1_on_turn = !match->local.p1_on_turn;
    break;
  case SINGLE:
    break;
  }
}

void start_turn(Match *match) {
  Player *player, *opponent;

  switch (match->mode) {
  case MULTI_REMOTE:
    player = match->remote.on_turn;
    opponent = get_opponent(match->player1, match->player2, player);
    send_only_type(player->client_fd, STR(WAIT_GUESS));
    send_only_type(opponent->client_fd, STR(WAIT_OPPONENT_GUESS));
    break;
  case MULTI_LOCAL:
    if (match->local.p1_on_turn) {
      send_only_type(match->player1->client_fd, STR(WAIT_GUESS));
    } else {
      send_only_type(match->player1->client_fd, STR(WAIT_OPPONENT_GUESS));
    }
    break;
  case SINGLE:
    send_only_type(match->player1->client_fd, STR(WAIT_GUESS));
    break;
  }

  if (match->turn_timer != NULL) {
    Timer_arm(match->turn_timer);
  }
}

TimerFireAction expire_turn_timer(TurnTimerData *timer_data) {
  printf("Turn timer expired!\n");
  GameServer *gs = timer_data->gs;
  Match *match = timer_data->match;
  Round *round = NULL;
  cJSON *guess_result_json = NULL;

  switch (match->mode) {
  case MULTI_REMOTE:
  case MULTI_LOCAL:
    swap_turn(match);
    break;
  case SINGLE:
    round = match->rounds[match->round_idx];

    for (size_t i = 0; i < round->wc_num; i++) {
      WordChallenge *wc = round->wc_list[i];
      for (size_t j = 0; j < match->word_len; j++) {
        wc->feedback[j] = LETTER_ABSENT;
      }
    }

    char *filler_guess = "";
    add_guess_attempt(round, filler_guess);
    guess_result_json = json_guess_result(filler_guess, round, match->word_len);
    send_json(match->player1->client_fd, guess_result_json);
    cJSON_Delete(guess_result_json);

    if (is_round_finished(round)) {
      return GS_end_round(gs, match);
    }
    break;
  }
  return TIMER_FIRE_REARM;
}

TimerFireAction expire_post_round_timer(Match *match) {}

bool is_round_finished(Round *round) {
  return (round->solved_num == round->wc_num) || (round->attempt_count >= round->max_attempts);
}

void add_guess_attempt(Round *round, char *guess) { round->guess_attempts[round->attempt_count++] = strdup(guess); }
