#include "game_types.h"
#include "game_logic.h"
#include "util.h"
#include <cjson/cJSON.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>

Match *new_match(size_t round_capacity, GameMode mode, size_t word_len) {
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

Round *new_round(WordChallenge *word_challenge, Player *starting_player) {
  Round *round = malloc(sizeof(Round));
  if (round == NULL) {
    perror("malloc");
    return NULL;
  }

  round->on_turn = starting_player;
  round->wc = word_challenge;

  return round;
}

void delete_round(Round *round) {
  delete_word_challenge(round->wc);
  free(round);
}

WordChallenge *new_word_challenge(int word_len, int max_attempts) {
  WordChallenge *wc = malloc(sizeof(WordChallenge));
  if (wc == NULL) {
    perror("malloc");
    return NULL;
  }

  wc->word_len = word_len;
  wc->attempt_count = 0;
  wc->max_attempts = max_attempts;
  wc->guess_attempts = malloc(sizeof(char *) * max_attempts);
  wc->word = get_random_word(word_len);
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

Player *new_player(Match *match, char *name) {
  Player *player = malloc(sizeof(Player));
  if (player == NULL) {
    perror("malloc");
    return NULL;
  }
  player->match = match;
  player->name = name;
  return player;
}

void delete_player(Player *player) {
  free(player);
  free(player->name);
}
