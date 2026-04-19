#include "timer.h"
#include <assert.h>
#include <stdio.h>
#include <unistd.h>

#include "test_util.h"

typedef struct TimerTestData {
  Timer *timer;
  int *counter;
} TimerTestData;

void test_timer_lifecycle(void);
void test_timer_list_examine(void);
void test_timer_rearm(void);
void test_timer_arm_within_examine(void);

static void toggle(bool *data);
static void increment(int *data);
static void increment_and_arm(TimerTestData *data);

void test_timer_lifecycle(void) {
  TimerList tl = {NULL};
  Timer *t1, *t2, *t3, *t4, *t_cur;
  bool toggle_switch = false;
  int counter = 0;

  t1 = new_timer(&tl, (TimerCallbackFunc)toggle, &toggle_switch, 30);
  t2 = new_timer(&tl, (TimerCallbackFunc)toggle, &toggle_switch, 10);
  t3 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 40);
  t4 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 5);

  Timer_arm(t1);
  Timer_arm(t2);
  Timer_arm(t3);
  Timer_arm(t4);

  Timer_fire(t1);
  assert(toggle_switch == true);
  Timer_fire(t2);
  assert(toggle_switch == false);
  Timer_fire(t3);
  assert(counter == 1);
  Timer_fire(t4);
  assert(counter == 2);
  Timer_fire(t4);
  assert(counter == 3);

  TimerList_print(&tl);

  t_cur = tl.head;
  assert(t_cur == t4);
  t_cur = t_cur->next;
  assert(t_cur == t2);
  t_cur = t_cur->next;
  assert(t_cur == t1);
  t_cur = t_cur->next;
  assert(t_cur == t3);

  Timer_disarm(t4);
  assert(tl.head == t2);
  Timer_disarm(t3);
  assert(tl.head == t2);
  Timer_disarm(t2);
  assert(tl.head == t1);
  Timer_disarm(t1);
  assert(tl.head == NULL);

  delete_timer(t1, false);
  delete_timer(t2, false);
  delete_timer(t3, false);
  delete_timer(t4, false);
}

void test_timer_list_examine(void) {
  TimerList tl = {NULL};
  Timer *t1, *t2, *t3, *t4;
  int counter = 0, sleep_seconds = 5;

  t1 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 3);
  t2 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 30);
  t3 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 15);
  t4 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 2);

  Timer_arm(t2);
  Timer_arm(t1);
  Timer_arm(t3);
  Timer_arm(t4);

  sleep(sleep_seconds);

  TimerList_examine(&tl);

  assert(counter == 2);
  assert(tl.head == t3);
  assert(tl.head->next == t2);

  TimerList_examine(&tl);

  assert(counter == 2);
  assert(tl.head == t3);
  assert(tl.head->next == t2);
}

void test_timer_rearm(void) {
  TimerList tl = {NULL};
  Timer *t1, *t2, *t3;
  int counter = 0, sleep_seconds = 3;

  t1 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 4);
  t2 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 2);
  t3 = new_timer(&tl, (TimerCallbackFunc)increment, &counter, 15);

  Timer_arm(t1);
  Timer_arm(t2);
  Timer_arm(t3);

  assert(tl.head == t2);
  TimerList_examine(&tl);
  assert(counter == 0);

  sleep_with_log(sleep_seconds);

  TimerList_examine(&tl);
  assert(counter == 1);

  Timer_arm(t2);
  assert(tl.head == t1);

  sleep_with_log(sleep_seconds);
  Timer_rearm(t2);

  TimerList_examine(&tl);
  assert(counter == 2);

  sleep_with_log(sleep_seconds);
  TimerList_examine(&tl);
  assert(counter == 3);
}

void test_timer_arm_within_examine(void) {
  TimerList tl = {NULL};
  Timer *t1, *t2, *t3, *t4;
  int counter = 0, sleep_seconds = 5;
  TimerTestData d1 = {NULL, &counter}, d2 = {NULL, &counter}, d3 = {NULL, &counter}, d4 = {NULL, &counter};

  t1 = new_timer(&tl, (TimerCallbackFunc)increment_and_arm, &d1, 1);
  d1.timer = t1;
  t2 = new_timer(&tl, (TimerCallbackFunc)increment_and_arm, &d2, 30);
  d2.timer = t2;
  t3 = new_timer(&tl, (TimerCallbackFunc)increment_and_arm, &d3, 7);
  d3.timer = t3;
  t4 = new_timer(&tl, (TimerCallbackFunc)increment_and_arm, &d4, 2);
  d4.timer = t4;

  Timer_arm(t2);
  Timer_arm(t1);
  Timer_arm(t3);
  Timer_arm(t4);

  printf("Pre-examine: ");
  TimerList_print(&tl);
  sleep(sleep_seconds);

  TimerList_examine(&tl);
  printf("Examine 1: ");
  TimerList_print(&tl);
  sleep(sleep_seconds);

  TimerList_examine(&tl);
  printf("Examine 2: ");
  TimerList_print(&tl);

  assert(counter = 5);
  assert(tl.head == t1);
  assert(tl.head->next == t4);
}

void increment_and_arm(TimerTestData *data) {
  Timer_disarm(data->timer);
  data->counter++;
  Timer_arm(data->timer);
}

void toggle(bool *data) { *data = !*data; }
void increment(int *data) { (*data)++; }
