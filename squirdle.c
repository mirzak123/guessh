#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#define WORD_LENGTH 5
#define WORDS_COUNT 2315
#define WORDS_FILE "words.txt"
#define MAX_INPUT 1000

#define ANSI_COLOR_RED "\x1b[31m"
#define ANSI_COLOR_GREEN "\x1b[32m"
#define ANSI_COLOR_YELLOW "\x1b[33m"
#define ANSI_COLOR_RESET "\x1b[0m"

char *rand_squirdle();
int rand_word_index();
int guess(char *);

int main() {
  srand(time(NULL));
  char *squirdle = rand_squirdle();

  printf("squirdle: %s\n", squirdle);

  if (guess(squirdle))
    printf(ANSI_COLOR_GREEN "SUCCESS! YOU GUESSED CORRECTLY" ANSI_COLOR_RESET
                            "\n");
  else
    printf(ANSI_COLOR_RED
           "OUT OF GUESSES! THE CORRECT WORD WAS: %s" ANSI_COLOR_RESET "\n",
           squirdle);

  free(squirdle);

  return 0;
}

char *rand_squirdle() {
  int fd, offset;
  char *squirdle;

  fd = open(WORDS_FILE, O_RDONLY);
  if (fd == -1) {
    perror("perror");
    exit(EXIT_FAILURE);
  }

  squirdle = malloc(WORD_LENGTH * sizeof(char));
  offset = rand_word_index();

  if ((pread(fd, squirdle, 5, offset)) == -1) {
    perror("pread");
    exit(EXIT_FAILURE);
  }

  return squirdle;
}

int rand_word_index() {
  int word_index = rand() % (WORDS_COUNT + 1);
  int byte_offset = (WORD_LENGTH + 1) * word_index;
  return byte_offset;
}

int guess(char *squirdle) {
  char *line = NULL;
  size_t size = 0;
  int guessed = 0;

  for (int i = 0; i < WORD_LENGTH + 1; i++) {
    printf("guess %d/%d: ", i + 1, WORD_LENGTH + 1);
    if (getline(&line, &size, stdin) == -1) {
      perror("getline");
      exit(EXIT_FAILURE);
    }

    if (strncmp(squirdle, line, WORD_LENGTH) == 0) {
      guessed = 1;
      break;
    }
  }

  free(line);
  return guessed;
}
