#ifndef GAME_TYPES_H
#define GAME_TYPES_H

typedef enum { TIE, PLAYER1_WINS, PLAYER2_WINS } Outcome;

typedef enum {
  SINGLE,
  MULTI_LOCAL,
  MULTI_REMOTE,
} GameMode;

typedef struct {
  int id;
  /* additional fields like name, socket_id, etc. */
} Player;

typedef struct {
  char *word;
  int word_len;
  int attempt_count; /* how many attempts have been made */
  int max_attempts;

  char **guess_attempts; /* array of all guess attempts made */
  int is_solved;         /* optional */
} WordChallenge;

typedef struct {
  /* When we turn this into a quordle-style game, we would store an array
   * of WordChallenge structs */
  WordChallenge *word;
  Outcome outcome;
  Player *starting_player;
} Round;

typedef struct {
  int round_count;    /* how many rounds have been played */
  int round_capacity; /* total amount of rounds */
  Round **rounds;
  GameMode mode;
  Outcome outcome;
  Player players[];
} Match;

#endif // !GAME_TYPES_H
