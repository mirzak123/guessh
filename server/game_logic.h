#ifndef GAME_LOGIC_H
#define GAME_LOGIC_H

#include <stdbool.h>
#include <stddef.h>

#define MIN_WORD_LEN 5
#define MAX_WORD_LEN 7

#define ENV_WORDS_PATH "WORDS_PATH"
#define DEFAULT_WORDS_PATH "./words"

#define FIVE_LETTER_WORD_FILE "five-letter.txt"
#define SIX_LETTER_WORD_FILE "six-letter.txt"
#define SEVEN_LETTER_WORD_FILE "seven-letter.txt"

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

bool evaluate_all_word_challenges(const char *guess_word, const char *target_word, LetterFeedback *feedback, size_t len);
bool evaluate_single_word_challenge(const char *guess_word, const char *target_word, LetterFeedback *feedback, size_t len);
char *get_random_word(WordStore *store);

#endif // !GAME_LOGIC_H
