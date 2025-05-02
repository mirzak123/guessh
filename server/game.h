#ifndef GAME_H
#define GAME_H

typedef enum {
  LETTER_ABSENT,
  LETTER_PRESENT,
  LETTER_CORRECT,
} LetterFeedback;

int evaluate_guess(const char *guess_word, const char *target_word, LetterFeedback *feedback, int len);
int contains(const char *target_word, char letter, int len);

#endif // !GAME_H
