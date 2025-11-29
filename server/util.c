#include "util.h"

// TODO: Generate uuid
int generate_unique_id(void) {
  static int id = 0;
  return ++id;
}
