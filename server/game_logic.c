#include "game_logic.h"
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/fcntl.h>
#include <unistd.h>

/* Returns 1 if correctly guessed, 0 otherwise. */
bool evaluate_guess(const char *guess_word, const char *target_word, LetterFeedback *feedback, int len) {
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

char *get_random_word(int word_len) {
  int fd, offset;
  char *word;

  WordStore *ws = &word_stores[word_len - MIN_WORD_LEN];
  printf("[get_random_word] getting word from file: %s\n", ws->file);

  fd = open(ws->file, O_RDONLY);
  if (fd == -1) {
    perror("open");
    exit(EXIT_FAILURE); // TODO: Maybe shouldn't kill program entirely
  }

  word = malloc(word_len * sizeof(char));
  offset = rand_word_index(ws);

  if ((pread(fd, word, word_len, offset)) == -1) {
    perror("pread");
    exit(EXIT_FAILURE);
  }

  return word;
}

int rand_word_index(WordStore *ws) {
  int word_index = rand() % (ws->word_count + 1);
  int byte_offset = (ws->word_len + 1) * word_index;
  return byte_offset;
}
