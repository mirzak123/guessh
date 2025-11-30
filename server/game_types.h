#ifndef GAME_TYPES_H
#define GAME_TYPES_H

#include <stddef.h>
typedef enum { TIE, PLAYER1_WINS, PLAYER2_WINS } Outcome;

typedef enum {
  SINGLE,
  MULTI_LOCAL,
  MULTI_REMOTE,
} GameMode;

// BUG: Identifying playrs by file descriptor will produce issues since file descriptors are
// reused after one player disconnects and another one joins.
// Ideally should create a system to interact with players only through playerId, and
// resolve the file descriptor at the last minute, when sending a message.
typedef struct {
  int fd;
  /* additional fields like name, socket_id, etc. */
} Player;

Player *new_player(int client_fd);
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
  Player *starting_player;
} Round;

Round *new_round(WordChallenge *word_challenge, Player *starting_player);
void delete_round(Round *round);

typedef struct Match {
  char *id;
  size_t round_current;  /* how many rounds have been played */
  size_t round_capacity; /* total amount of rounds */
  Round **rounds;
  GameMode mode;
  Outcome outcome;
  Player *player1;
  Player *player2;
  struct Match *next;
} Match;

Match *new_match(size_t round_capacity, GameMode mode);
void Match_add_player(Match *match, int client_fd);
void delete_match(Match *match);

#endif // !GAME_TYPES_H
