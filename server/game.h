#ifndef GAME_H
#define GAME_H

#define WORDS_FILE "../server/words.txt"

typedef enum {
  LETTER_ABSENT,
  LETTER_PRESENT,
  LETTER_CORRECT,
} LetterFeedback;

int evaluate_guess(const char *guess_word, const char *target_word, LetterFeedback *feedback, int len);
void get_random_word(char *str);

#endif // !GAME_H
