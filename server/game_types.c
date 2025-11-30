#include "game_types.h"
#include "game_logic.h"
#include "game_server.h"
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
  match->round_current = 0;
  match->rounds = malloc(sizeof(Round) * round_capacity);
  match->word_len = word_len;

  return match;
}

void Match_add_player(Match *match, int client_fd) {
  Player *player = new_player(client_fd);
  if (player == NULL) {
    printf("[Match_add_player] error: new_player() returned NULL\n");
    return;
  }

  if (match->player1 == NULL) { // first player
    match->player1 = player;
    if (match->mode == SINGLE) {
      GS_start_match(match);
    }
  } else if (match->player2 == NULL) { // second player
    if (match->mode == SINGLE) {
      printf("[Match_add_player] error: trying to add second player to a match in SINGLE mode\n");
      return;
    } else { // multiplayer
      match->player2 = player;
      // TODO: start match
    }
  } else {
    printf("[Match_add_player] error: trying to add a player to a match that has 2 players\n");
    return;
  }
}

Round *new_round(WordChallenge *word_challenge, Player *starting_player) {
  Round *round = malloc(sizeof(Round));
  if (round == NULL) {
    perror("malloc");
    return NULL;
  }

  round->starting_player = starting_player;
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
  wc->guess_attempts = malloc(sizeof(char) * word_len * max_attempts);
  wc->word = get_random_word(word_len);

  return wc;
}

void delete_word_challenge(WordChallenge *wc) {
  free(wc->word);

  for (int i = 0; i < (int)wc->attempt_count; i++) {
    free(wc->guess_attempts[i]);
  }
  free(wc->guess_attempts);
  free(wc);
}

Player *new_player(int client_fd) {
  Player *player = malloc(sizeof(Player));
  if (player == NULL) {
    perror("malloc");
    return NULL;
  }

  player->fd = client_fd;
  return player;
}

void delete_player(Player *player) { free(player); }
