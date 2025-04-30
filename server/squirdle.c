#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#define WORD_LENGTH 5
#define WORDS_COUNT 2315
#define WORDS_FILE "../server/words.txt"
#define MAX_INPUT 1000

#define ANSI_COLOR_RED "\x1b[31m"
#define ANSI_COLOR_GREEN "\x1b[32m"
#define ANSI_COLOR_YELLOW "\x1b[33m"
#define ANSI_COLOR_RESET "\x1b[0m"
#define ANSI_MOVE_UP_LINE "\033[A"
#define ANSI_CLEAR_LINE "\033[K"

char *rand_squirdle();
int rand_word_index();
int guess(char *);
int guess_results(char *, char *, int);
void draw_guess_prompt(int);

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
    perror("open");
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
    draw_guess_prompt(i);
    int read_bytes;
    // FIX: read_bytes should be used somehow to prevent a newline if user
    // inputs a shorter word
    if ((read_bytes = getline(&line, &size, stdin)) == -1) {
      perror("getline");
      exit(EXIT_FAILURE);
    }

    printf(ANSI_MOVE_UP_LINE "\r" ANSI_CLEAR_LINE);
    draw_guess_prompt(i);

    int correct = guess_results(squirdle, line, WORD_LENGTH);

    if (correct) {
      guessed = 1;
      break;
    }
  }

  free(line);
  return guessed;
}

int guess_results(char *squirdle, char *guess, int len) {
  int total_correct = 0;

  for (int i = 0; i < len; i++) {
    if (squirdle[i] == guess[i]) {
      printf(ANSI_COLOR_GREEN "%c" ANSI_COLOR_RESET, guess[i]);
      total_correct++;
    } else {
      printf("%c", guess[i]);
    }
  }
  printf("\n");
  return total_correct == len;
}

void draw_guess_prompt(int i) {
  printf("guess %d/%d: ", i + 1, WORD_LENGTH + 1);
}
