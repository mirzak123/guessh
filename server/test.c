#include "game_logic.c"
#include "game_types.h"
#include "hash_table.h"
#include "util.h"
#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define WORD_LEN 5

void test_evaluate_guess(void);
void test_hash_table(void);

void assert_feedback(LetterFeedback *feedback, LetterFeedback *expected);
void print_feedback(LetterFeedback *feedback);
void test_generate_random_string(void);
void test_call_HT_delete_on_empty_hash_table(void);

int main(void) {
  srand(time(NULL));
  test_evaluate_guess();
  test_hash_table();
  test_call_HT_delete_on_empty_hash_table();

  printf("All tests passed!\n");
  return 0;
}

void test_generate_random_string(void) {
  char *s = generate_random_string(7);
  printf("%lu: |%s|\n", strlen(s), s);
  free(s);
}

void test_evaluate_guess(void) {
  LetterFeedback feedback[WORD_LEN];
  int r;

  WordChallenge *wc = &(WordChallenge){
      NULL,
      WORD_LEN,
      OUTCOME_NONE,
      feedback,
  };

  wc->word = "ocean";
  r = evaluate_word_challenge_guess("ocean", wc);
  assert_feedback(feedback, (LetterFeedback[]){2, 2, 2, 2, 2});
  assert(r == 1);

  wc->word = "cubic";
  r = evaluate_word_challenge_guess("cubic", wc);
  assert_feedback(feedback, (LetterFeedback[]){2, 2, 2, 2, 2});
  assert(r == 1);

  wc->word = "piles";
  r = evaluate_word_challenge_guess("pulls", wc);
  assert_feedback(feedback, (LetterFeedback[]){2, 0, 2, 0, 2});
  assert(r == 0);

  wc->word = "leaky";
  evaluate_word_challenge_guess("pulls", wc);
  assert_feedback(feedback, (LetterFeedback[]){0, 0, 1, 0, 0});
  assert(r == 0);

  wc->word = "whose";
  r = evaluate_word_challenge_guess("echos", wc);
  assert_feedback(feedback, (LetterFeedback[]){1, 0, 1, 1, 1});
  assert(r == 0);

  wc->word = "whose";
  r = evaluate_word_challenge_guess("shoes", wc);
  assert_feedback(feedback, (LetterFeedback[]){1, 2, 2, 1, 0});
  assert(r == 0);

  wc->word = "cubic";
  r = evaluate_word_challenge_guess("lucid", wc);
  assert_feedback(feedback, (LetterFeedback[]){0, 2, 1, 2, 0});
  assert(r == 0);

  wc->word = "lucid";
  r = evaluate_word_challenge_guess("cubic", wc);
  assert_feedback(feedback, (LetterFeedback[]){1, 2, 0, 2, 0});
  assert(r == 0);

  wc->word = "lilac";
  r = evaluate_word_challenge_guess("spill", wc);
  assert_feedback(feedback, (LetterFeedback[]){0, 0, 1, 1, 1});
  assert(r == 0);

  wc->word = "spill";
  r = evaluate_word_challenge_guess("lilac", wc);
  assert_feedback(feedback, (LetterFeedback[]){1, 1, 1, 0, 0});
  assert(r == 0);

  wc->word = "tutti";
  r = evaluate_word_challenge_guess("totem", wc);
  assert_feedback(feedback, (LetterFeedback[]){2, 0, 2, 0, 0});
  assert(r == 0);

  wc->word = "totem";
  r = evaluate_word_challenge_guess("tutti", wc);
  assert_feedback(feedback, (LetterFeedback[]){2, 0, 2, 0, 0});
  assert(r == 0);
}

void test_hash_table(void) {
  Value value;
  char *value_str = "Hello world";
  int value_int = 1000, count = 0;

  HashTable *table = HT_create();
  assert(table != NULL);

  HT_set(table, KEY("key"), &value_int);
  count++;
  value = HT_get((table), KEY("key"));
  assert(*(int *)value == value_int);
  assert(table->capacity == 8);

  for (int i = count + 1; i < 10; i++) {
    HT_set(table, KEY(i), value_str);
    count++;
    value = HT_get(table, KEY(i));
    assert(!strcmp((char *)value, value_str));

    if (i <= 6)
      assert(table->capacity == 8);
    else
      assert(table->capacity == 16);
  }

  int x = 7;
  HT_set(table, KEY(x), value_str);
  assert(value_str == HT_get(table, KEY(x)));
  HT_delete(table, KEY(x));
  assert(NULL == HT_get(table, KEY(x)));

  // Find nonexistent
  value = HT_get(table, KEY("nonexistent"));
  assert(value == NULL);
  assert(table->capacity == 16);

  HT_destroy(table, NULL);
}

void test_call_HT_delete_on_empty_hash_table(void) {
  HashTable *ht = HT_create();
  HT_delete(ht, KEY(1));
}

void assert_feedback(LetterFeedback *feedback, LetterFeedback *expected) {
  for (int i = 0; i < WORD_LEN; i++) {
    assert(feedback[i] == expected[i]);
  }
}

void print_feedback(LetterFeedback *feedback) {
  printf("{");
  for (int i = 0; i < WORD_LEN; i++) {
    printf("%d ", feedback[i]);
  }
  printf("}\n");
}
