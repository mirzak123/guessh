#include "game_logic.h"
#include "util.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define WORD_LEN 5

void test_evaluate_guess(void);
void assert_feedback(LetterFeedback *feedback, LetterFeedback *expected);
void print_feedback(LetterFeedback *feedback);
void test_generate_random_string(void);

int main(void) {
  srand(time(NULL));
  test_evaluate_guess();

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

  r = evaluate_guess("ocean", "ocean", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){2, 2, 2, 2, 2});
  assert(r == 1);

  r = evaluate_guess("cubic", "cubic", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){2, 2, 2, 2, 2});
  assert(r == 1);

  r = evaluate_guess("pulls", "piles", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){2, 0, 2, 0, 2});
  assert(r == 0);

  evaluate_guess("pulls", "leaky", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){0, 0, 1, 0, 0});
  assert(r == 0);

  r = evaluate_guess("echos", "whose", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){1, 0, 1, 1, 1});
  assert(r == 0);

  r = evaluate_guess("shoes", "whose", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){1, 2, 2, 1, 0});
  assert(r == 0);

  r = evaluate_guess("lucid", "cubic", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){0, 2, 1, 2, 0});
  assert(r == 0);

  r = evaluate_guess("cubic", "lucid", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){1, 2, 0, 2, 0});
  assert(r == 0);

  r = evaluate_guess("spill", "lilac", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){0, 0, 1, 1, 1});
  assert(r == 0);

  r = evaluate_guess("lilac", "spill", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){1, 1, 1, 0, 0});
  assert(r == 0);

  r = evaluate_guess("totem", "tutti", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){2, 0, 2, 0, 0});
  assert(r == 0);

  r = evaluate_guess("tutti", "totem", feedback, WORD_LEN);
  assert_feedback(feedback, (LetterFeedback[]){2, 0, 2, 0, 0});
  assert(r == 0);
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
