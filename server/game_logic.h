#ifndef GAME_LOGIC_H
#define GAME_LOGIC_H

#include <stdbool.h>
#include <stddef.h>

#define MIN_WORD_LEN 5
#define MAX_WORD_LEN 7

#define ENV_WORDS_PATH "WORDS_PATH"
#define DEFAULT_WORDS_PATH "./words"

#define FIVE_LETTER_WORD_FILE "five-letter-secret.txt"
#define SIX_LETTER_WORD_FILE "six-letter-secret.txt"
#define SEVEN_LETTER_WORD_FILE "seven-letter-secret.txt"

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
char *get_random_word(WordStore *store);

#endif // !GAME_LOGIC_H
