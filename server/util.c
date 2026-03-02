#include "util.h"
#include <stdlib.h>

// TODO: Generate uuid
int generate_unique_id(void) {
  static int id = 0;
  return ++id;
}

char *generate_random_string(int len) {
  char *str = malloc(sizeof(char) * (len + 1));
  for (int i = 0; i < len; i++) {
    str[i] = rand() % ('Z' + 1 - 'A') + 'A';
  }
  str[len] = '\0';
  return str;
}
