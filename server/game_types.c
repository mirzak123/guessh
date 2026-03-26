#include "game_types.h"
#include "game_logic.h"
#include "room.h"
#include "timer.h"
#include "util.h"

#include <assert.h>
#include <cjson/cJSON.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

Match *new_match(GameMode mode, GameFormat format, size_t round_capacity, size_t word_len) {
  Match *match = calloc(1, sizeof(Match));
  if (match == NULL) {
    perror("calloc");
    return NULL;
  }

  match->id = malloc(16);
  snprintf(match->id, 16, "%d", generate_unique_id());

  match->round_capacity = round_capacity;
  match->mode = mode;
  match->format = format;
  match->word_len = word_len;
  match->round_idx = -1;
  match->turn_timer = NULL;

  match->rounds = calloc(round_capacity, sizeof(Round *));
  if (match->rounds == NULL) {
    perror("calloc rounds");
    free(match->id);
    free(match);
    return NULL;
  }

  return match;
}

void delete_match(Match *match) {
  assert(match != NULL);
  printf("Deleting match [id: %s]\n", match->id);

  free(match->id);
  free(match->room_id);

  for (int i = 0; i <= match->round_idx; i++) {
    printf("Deleting round idx: %d\n", i);
    delete_round(match->rounds[i]);
  }
  free(match->rounds);

  if (match->turn_timer != NULL) {
    delete_timer(match->turn_timer, true);
  }

  free(match);
}

Round *new_round(WordChallenge **word_challenges, size_t wc_num, size_t max_attempts) {
  Round *round = calloc(1, sizeof(Round));
  if (round == NULL) {
    perror("calloc");
    return NULL;
  }

  round->wc_list = word_challenges;
  round->wc_num = wc_num;
  round->solved_num = 0;

  round->attempt_count = 0;
  round->max_attempts = max_attempts;
  round->guess_attempts = calloc(max_attempts, sizeof(char *));
  if (round->guess_attempts == NULL) {
    perror("calloc guess_attempts");
    return NULL;
  }

  return round;
}

void delete_round(Round *round) {
  for (size_t i = 0; i < round->wc_num; i++) {
    delete_word_challenge(round->wc_list[i]);
  }

  for (int i = 0; i < (int)round->attempt_count; i++) {
    free(round->guess_attempts[i]);
  }

  free(round->guess_attempts);
  free(round->wc_list);
  free(round);
}

WordChallenge *new_word_challenge(WordStore *store) {
  WordChallenge *wc = calloc(1, sizeof(WordChallenge));
  if (wc == NULL) {
    perror("calloc");
    return NULL;
  }

  wc->word = get_random_word(store);
  printf("[new_word_challenge] word: %s\n", wc->word);
  wc->solved_by = OUTCOME_NONE;
  wc->len = store->word_len;

  wc->feedback = calloc(wc->len, sizeof(LetterFeedback));
  if (wc->feedback == NULL) {
    perror("wc->feedback calloc");
    return NULL;
  }

  return wc;
}

void delete_word_challenge(WordChallenge *wc) {
  free(wc->feedback);
  free(wc);
}

Player *new_player(int client_fd, char *name) {
  Player *player = calloc(1, sizeof(Player));
  if (player == NULL) {
    perror("calloc");
    return NULL;
  }

  player->client_fd = client_fd;
  player->match = NULL;
  player->wants_rematch = false;

  if (name != NULL) {
    player->name = strdup(name);
  } else {
    player->name = NULL;
  }

  return player;
}

void delete_player(Player *player) {
  if (player->match) {
    Match *match = player->match;
    if (player == match->player1) {
      match->player1 = NULL;
    } else if (player == match->player2) {
      match->player2 = NULL;
    }
  }

  if (player->room) {
    Room *room = player->room;
    if (player == room->player1) {
      room->player1 = NULL;
    } else if (player == room->player2) {
      room->player2 = NULL;
    }
  }

  free(player->name);
  free(player);
}
