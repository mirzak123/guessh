#ifndef GAME_TYPES_H
#define GAME_TYPES_H

#include "game_logic.h"
#include "timer.h"
#include <stdbool.h>
#include <stddef.h>

#define MAX_CLIENT_DATA 1024

struct Room;
struct WordChallenge;

typedef enum { OUTCOME_NONE, OUTCOME_PLAYER1, OUTCOME_PLAYER2 } Outcome;

typedef enum {
  SINGLE,
  MULTI_LOCAL,
  MULTI_REMOTE,
} GameMode;

typedef enum {
  WORDLE,
  QUORDLE,
} GameFormat;

typedef struct Player {
  int client_fd;
  char *name;
  struct Match *match;
  struct Room *room;
  bool ready_next_round;
  bool wants_rematch;
} Player;

Player *new_player(int client_fd, char *name);
void delete_player(Player *player);

typedef struct WordChallenge {
  char *word;
  size_t len;
  Outcome solved_by;
  LetterFeedback *feedback;
} WordChallenge;

WordChallenge *new_word_challenge(WordStore *store);
void delete_word_challenge(WordChallenge *wc);

typedef struct Round {
  WordChallenge **wc_list;
  size_t wc_num;
  size_t solved_num;

  size_t attempt_idx;
  size_t max_attempts;

  char **guess_attempts;
  int points;
} Round;

Round *new_round(WordChallenge **word_challenges, size_t wc_num, size_t max_attempts);
void delete_round(Round *round);

typedef struct Match {
  char *id;
  bool is_active;
  int round_idx;
  size_t word_len;
  size_t round_capacity;
  char *room_id;
  Round **rounds;
  GameMode mode;
  GameFormat format;
  Outcome outcome;
  Player *player1;
  Player *player2;
  Timer *turn_timer;
  Timer *post_round_timer;
  union {
    struct {
      bool p1_on_turn;
      bool p1_start_match;
      bool p1_start_round;
    } local;
    struct {
      Player *on_turn;
      Player *match_starter;
      Player *round_starter;
    } remote;
  };
} Match;

Match *new_match(GameMode mode, GameFormat format, size_t round_capacity, size_t word_len);
void Match_start_match(Match *match);
void Match_start_round(Match *match);
void delete_match(Match *match);

#endif // !GAME_TYPES_H
