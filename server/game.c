#include "game.h"

int evaluate_guess(const char *guess_word, const char *target_word, LetterFeedback *feedback, int len) {
  int alphabet[26] = {0};
  int correct_count = 0;

  /* 1st pass: LETTER_CORRECT */
  for (int i = 0; i < len; i++) {
    if (guess_word[i] == target_word[i]) {
      feedback[i] = LETTER_CORRECT;
      correct_count++;
    } else {
      alphabet[target_word[i] - 'a']++; /* count letter */
      feedback[i] = LETTER_ABSENT;      /* clear feedback array */
    }
  }

  /* all letters in correct position */
  if (correct_count == len)
    return 1;

  /* 2nd pass: LETTER_PRESENT */
  for (int i = 0; i < len; i++) {
    if (feedback[i] != LETTER_CORRECT && alphabet[guess_word[i] - 'a'] > 0) {
      feedback[i] = LETTER_PRESENT;
      alphabet[guess_word[i] - 'a']--;
    }
  }

  return 0;
}
