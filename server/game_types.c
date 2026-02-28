#include "game_types.h"
#include "game_logic.h"
#include "util.h"
#include <_string.h>
#include <cjson/cJSON.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>

Match *new_match(GameMode mode, size_t round_capacity, size_t word_len) {
  Match *match = malloc(sizeof(Match));
  if (match == NULL) {
    perror("malloc");
    return NULL;
  }

  match->id = malloc(sizeof(long));
  sprintf(match->id, "%d", generate_unique_id());

  match->round_capacity = round_capacity;
  match->mode = mode;
  match->round_idx = -1;
  match->rounds = malloc(sizeof(Round) * round_capacity);
  match->word_len = word_len;

  return match;
}

void delete_match(Match *match) { // TODO: Call this somewhere. Currently no match is ever deleted.
  free(match->id);
  free(match->room_id);
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

Round *new_round(WordChallenge *word_challenge) {
  Round *round = malloc(sizeof(Round));
  if (round == NULL) {
    perror("malloc");
    return NULL;
  }

  round->wc = word_challenge;

  return round;
}

void delete_round(Round *round) {
  delete_word_challenge(round->wc);
  free(round);
}

WordChallenge *new_word_challenge(WordStore *store, int max_attempts) {
  WordChallenge *wc = malloc(sizeof(WordChallenge));
  if (wc == NULL) {
    perror("malloc");
    return NULL;
  }

  wc->word_len = store->word_len;
  wc->attempt_count = 0;
  wc->max_attempts = max_attempts;
  wc->guess_attempts = malloc(sizeof(char *) * max_attempts);
  wc->word = get_random_word(store);
  printf("[new_word_challenge] word: %s\n", wc->word);

  return wc;
}

void delete_word_challenge(WordChallenge *wc) {
  free(wc->word);

  printf("Deleting word challenge. Attempt count: %d\n", (int)wc->attempt_count);
  for (int i = 0; i < (int)wc->attempt_count; i++) {
    free(wc->guess_attempts[i]);
  }
  free(wc->guess_attempts);
  free(wc);
}

Player *new_player(int client_fd, char *name) {
  Player *player = malloc(sizeof(Player));
  if (player == NULL) {
    perror("malloc");
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
  free(player->name);
  free(player);
}
