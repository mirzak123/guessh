#include "game_logic.c"
#include "game_types.h"
#include <assert.h>

#define WORD_LEN 5

static void test_evaluate_word_challenge_guess(void);
static void assert_feedback(LetterFeedback *feedback, LetterFeedback *expected);

void test_evaluate_word_challenge_guess(void) {
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

void assert_feedback(LetterFeedback *feedback, LetterFeedback *expected) {
  for (int i = 0; i < WORD_LEN; i++) {
    assert(feedback[i] == expected[i]);
  }
}
