#ifndef GAME_LOGIC_H
#define GAME_LOGIC_H

#include <stddef.h>
#define WORD_STORE_LEN 3
#define MIN_WORD_LEN 5
#define MAX_WORD_LEN 7

typedef struct {
  char *file;
  size_t word_len;
  size_t word_count;
} WordStore;

static WordStore word_stores[] = {
    {"./words/five-letter.txt", 5, 1168},
    {"./words/six-letter.txt", 6, 1164},
    {"./words/seven-letter.txt", 7, 1164},
};

typedef enum {
  LETTER_ABSENT,
  LETTER_PRESENT,
  LETTER_CORRECT,
} LetterFeedback;

int evaluate_guess(const char *guess_word, const char *target_word, LetterFeedback *feedback, int len);
char *get_random_word(int word_len);
int rand_word_index(WordStore *ws); // TODO:  Remove from game_logic.h -- should be private

#endif // !GAME_LOGIC_H
