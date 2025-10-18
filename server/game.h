#ifndef GAME_H
#define GAME_H

#define WORDS_FILE "./words/five-letter.txt"
#define WORD_LENGTH 5
#define WORD_COUNT 1168

typedef enum {
  LETTER_ABSENT,
  LETTER_PRESENT,
  LETTER_CORRECT,
} LetterFeedback;

int evaluate_guess(const char *guess_word, const char *target_word, LetterFeedback *feedback, int len);
char *get_random_word();
int rand_word_index(); // TODO:  Remove from game.h -- should be private

#endif // !GAME_H
