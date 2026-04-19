#include "util_test.h"
#include "../src/util.h"

#include <assert.h>
#include <stdlib.h>
#include <string.h>

void test_generate_random_string(void) {
  char *s = generate_random_string(7);
  assert(strlen(s) == 7);
  free(s);
}
