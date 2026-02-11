#ifndef GAME_TYPES_H
#define GAME_TYPES_H

#include <stdbool.h>
#include <stddef.h>

#define MAX_CLIENT_DATA 1024

typedef enum { TIE, PLAYER1_WINS, PLAYER2_WINS } Outcome;

typedef enum {
  SINGLE,
  MULTI_LOCAL,
  MULTI_REMOTE,
} GameMode;

typedef struct {
  struct Match *match;
  char *name;
  int client_fd;
} Player;

Player *new_player(struct Match *match, char *name);
void delete_player(Player *player);

typedef struct {
  char *word;
  size_t word_len;
  size_t attempt_count; /* how many attempts have been made */
  size_t max_attempts;

  char **guess_attempts; /* array of all guess attempts made */
  int is_solved;         /* optional */
} WordChallenge;

WordChallenge *new_word_challenge(int word_len, int max_attempts);
void delete_word_challenge(WordChallenge *word_challenge);

typedef struct {
  /* When we turn this into a quordle-style game, we would store an array
   * of WordChallenge structs */
  WordChallenge *wc;
  Outcome outcome;
  Player *on_turn;
} Round;

Round *new_round(WordChallenge *word_challenge, Player *starting_player);
void delete_round(Round *round);

typedef struct Match {
  char *id; // TODO: Remove: Unused
  int round_idx;
  size_t round_capacity; /* total amount of rounds */
  Round **rounds;
  GameMode mode;
  Outcome outcome;
  Player *player1;
  Player *player2;
  size_t word_len;
  struct Match *next;
} Match;

Match *new_match(size_t round_capacity, GameMode mode, size_t word_len);
void Match_start_match(Match *match);
void Match_start_round(Match *match);
void delete_match(Match *match);

#endif // !GAME_TYPES_H
