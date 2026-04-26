#ifndef GAME_LOGIC_H
#define GAME_LOGIC_H

#include <stdbool.h>
#include <stddef.h>

#define MIN_WORD_LEN 5
#define MAX_WORD_LEN 7

#define ENV_WORDS_PATH "WORDS_PATH"
#define DEFAULT_WORDS_PATH "./words"

#define FIVE_LETTER_SECRET_WORD_FILE "five-letter-secret.txt"
#define SIX_LETTER_SECRET_WORD_FILE "six-letter-secret.txt"
#define SEVEN_LETTER_SECRET_WORD_FILE "seven-letter-secret.txt"

#define FIVE_LETTER_VALID_WORD_FILE "five-letter-valid.txt"
#define SIX_LETTER_VALID_WORD_FILE "six-letter-secret.txt"     // TODO: Valid word set not yet constructed
#define SEVEN_LETTER_VALID_WORD_FILE "seven-letter-secret.txt" // TODO: Valid word set not yet constructed

struct WordChallenge;

typedef enum {
  LETTER_ABSENT,
  LETTER_PRESENT,
  LETTER_CORRECT,
} LetterFeedback;

typedef struct {
  char **words;
  size_t word_len;
  size_t word_count;
} WordStore;

WordStore *new_word_store(char *filename, size_t word_len);
void delete_word_store(WordStore *store);

size_t evaluate_guess(const char *guess, struct WordChallenge **wc_list, size_t wc_num, bool player1_on_turn);
bool evaluate_word_challenge_guess(const char *guess, struct WordChallenge *wc);
char *get_random_word(WordStore *store);
bool is_valid_guess(WordStore *store, char *guess);

#endif // !GAME_LOGIC_H
