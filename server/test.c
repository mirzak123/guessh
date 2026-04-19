#include <stdio.h>
#include <stdlib.h>
#include <time.h>

#include "game_logic_test.c"
#include "hash_table_test.c"
#include "timer_test.c"
#include "util_test.c"

static void timer_test(void);
static void hash_table_test(void);
static void game_logic_test(void);
static void util_test(void);

int main(void) {
  srand(time(NULL));

  Test tests[] = {timer_test, hash_table_test, game_logic_test, util_test};
  int n = sizeof(tests) / sizeof(tests[0]);

  printf("Starting tests...\n");
  for (int i = 0; i < n; i++) {
    tests[i]();
  }
  printf("All tests passed!\n");

  return 0;
}

void timer_test(void) {
  test_timer_lifecycle();
  test_timer_list_examine();
  test_timer_rearm();
  test_timer_arm_within_examine();
}

void hash_table_test(void) {
  test_hash_table();
  test_call_HT_delete_on_empty_hash_table();
}

void game_logic_test(void) { test_evaluate_word_challenge_guess(); }

void util_test(void) { test_generate_random_string(); }
