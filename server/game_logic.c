#include "game_logic.h"
#include "game_types.h"
#include <fcntl.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

static bool evaluate_word_challenge_guess(const char *guess, WordChallenge *wc);
static char **read_fixed_lines(FILE *file, size_t line_len, size_t *out_count);

size_t evaluate_guess(const char *guess, WordChallenge **wc_list, size_t wc_num, bool player1_on_turn) {
  size_t solve_count = 0;

  for (size_t i = 0; i < wc_num; i++) {
    WordChallenge *wc = wc_list[i];

    bool solved = evaluate_word_challenge_guess(guess, wc);
    if (solved) {
      solve_count++;
      wc->solved_by = player1_on_turn ? OUTCOME_PLAYER1 : OUTCOME_PLAYER2;
    }
  }

  return solve_count;
}

/* Returns 1 if correctly guessed, 0 otherwise. */
bool evaluate_word_challenge_guess(const char *guess, WordChallenge *wc) {
  size_t alphabet[26] = {0};
  size_t correct_count = 0;

  /* 1st pass: LETTER_CORRECT */
  for (size_t i = 0; i < wc->len; i++) {
    if (guess[i] == wc->word[i]) {
      wc->feedback[i] = LETTER_CORRECT;
      correct_count++;
    } else {
      alphabet[wc->word[i] - 'a']++;   /* count letter */
      wc->feedback[i] = LETTER_ABSENT; /* clear feedback array */
    }
  }

  /* all letters in correct position */
  if (correct_count == wc->len)
    return true;

  /* 2nd pass: LETTER_PRESENT */
  for (size_t i = 0; i < wc->len; i++) {
    if (wc->feedback[i] != LETTER_CORRECT && alphabet[guess[i] - 'a'] > 0) {
      wc->feedback[i] = LETTER_PRESENT;
      alphabet[guess[i] - 'a']--;
    }
  }

  return false;
}

char *get_random_word(WordStore *store) {
  int index = rand() % (store->word_count + 1);
  return store->words[index];
}

WordStore *new_word_store(char *filename, size_t word_len) {
  WordStore *store = malloc(sizeof(WordStore));
  if (store == NULL) {
    perror("malloc");
    exit(EXIT_FAILURE);
  }

  char *words_base = getenv(ENV_WORDS_PATH);
  if (words_base == NULL) {
    words_base = DEFAULT_WORDS_PATH;
  }

  char filepath[512];
  snprintf(filepath, sizeof(filepath), "%s/%s", words_base, filename);

  FILE *file = fopen(filepath, "r");
  if (file == NULL) {
    perror("open");
    exit(EXIT_FAILURE);
  }

  store->word_len = word_len;
  store->words = read_fixed_lines(file, word_len, &store->word_count);

  if (fclose(file) == EOF) {
    perror("fclose");
    exit(EXIT_FAILURE);
  }

  return store;
}

void delete_word_store(WordStore *store) {
  for (size_t i = 0; i < store->word_count; i++) {
    free(store->words[i]);
  }
  free(store->words);
  free(store);
}

static char **read_fixed_lines(FILE *file, size_t line_len, size_t *out_count) {
  int capacity = 0;
  int count = 0;
  char **lines = NULL;

  for (;;) {
    char *line = malloc(line_len + 1);
    if (line == NULL) {
      perror("malloc");
      exit(EXIT_FAILURE);
    }

    size_t bytes = fread(line, 1, line_len, file);
    if (bytes != line_len) {
      free(line);
      break;
    }

    line[line_len] = '\0';
    fgetc(file);

    if (count == capacity) {
      capacity = capacity ? capacity * 2 : 16;
      void *tmp = realloc(lines, capacity * sizeof(char *));
      if (tmp == NULL) {
        free(lines);
        perror("realloc");
        exit(EXIT_FAILURE);
      }
      lines = tmp;
    }

    lines[count++] = line;
  }

  *out_count = count;
  return lines;
}
