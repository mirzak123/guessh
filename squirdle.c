#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>

#define WORD_LENGTH 5
#define WORDS_COUNT 2315
#define WORDS_FILE "words.txt"
#define TRIES 5
#define MAX_INPUT 1000

char *rand_squirdle();
int rand_word_index();
int guess(char *);

int main() {
  srand(time(NULL));
  char *squirdle = rand_squirdle();

  printf("squirdle: %s\n", squirdle);
  guess(squirdle);
  free(squirdle);

  return 0;
}

char *rand_squirdle() {
  int file = open(WORDS_FILE, O_RDONLY);
  char *squirdle = malloc(WORD_LENGTH * sizeof(char));
  int offset = rand_word_index();

  pread(file, squirdle, 5, offset);

  return squirdle;
}

int rand_word_index() {
  int word_index = rand() % (WORDS_COUNT + 1);
  int byte_offset = (WORD_LENGTH + 1) * word_index;
  return byte_offset;
}

int guess(char *squirdle) {
  char *buf[MAX_INPUT];
  size_t max_input = MAX_INPUT;
  for (int i = 0; i < TRIES; i++) {
    getline(buf, &max_input, stdin);
    printf("guess %d: %s", i + 1, *buf);
  }
  return 0;
}
