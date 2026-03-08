#ifndef GAME_TYPES_H
#define GAME_TYPES_H

#include "game_logic.h"
#include <stdbool.h>
#include <stddef.h>

#define MAX_CLIENT_DATA 1024

struct Room;

typedef enum { OUTCOME_NONE, OUTCOME_PLAYER1, OUTCOME_PLAYER2 } Outcome;

typedef enum {
  SINGLE,
  MULTI_LOCAL,
  MULTI_REMOTE,
} GameMode;

typedef struct Player {
  int client_fd;
  char *name;
  struct Match *match;
  struct Room *room;
  bool wants_rematch;
} Player;

Player *new_player(int client_fd, char *name);
void delete_player(Player *player);

typedef struct WordChallenge {
  char *word;
  size_t word_len;
  size_t attempt_count; /* how many attempts have been made */
  size_t max_attempts;

  char **guess_attempts; /* array of all guess attempts made */
  int is_solved;         /* optional */
} WordChallenge;

WordChallenge *new_word_challenge(WordStore *store, int max_attempts);
void delete_word_challenge(WordChallenge *word_challenge);

typedef struct Round {
  /* When we turn this into a quordle-style game, we would store an array
   * of WordChallenge structs */
  WordChallenge *wc;
  Outcome outcome;
} Round;

Round *new_round(WordChallenge *word_challenge);
void delete_round(Round *round);

typedef struct Match {
  char *id;
  int round_idx;
  size_t word_len;
  size_t round_capacity; /* total amount of rounds */
  char *room_id;
  Round **rounds;
  GameMode mode;
  Outcome outcome;
  Player *player1;
  Player *player2;
  union {
    struct {
      bool player1_on_turn;
      bool player1_started_round;
    } local;
    struct {
      Player *on_turn;
      Player *round_starter;
    } remote;
  };
} Match;

Match *new_match(GameMode mode, size_t round_capacity, size_t word_len);
void Match_start_match(Match *match);
void Match_start_round(Match *match);
void delete_match(Match *match);

#endif // !GAME_TYPES_H
