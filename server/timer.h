#ifndef TIMER_H
#define TIMER_H

#include <stdbool.h>
#include <stddef.h>

typedef enum { TIMER_FIRE_NONE, TIMER_FIRE_REARM } TimerFireAction;

typedef TimerFireAction (*TimerCallbackFunc)(void *data);
typedef void *TimerCallbackData;

typedef struct Timer {
  int id;
  long timestamp;
  size_t seconds;
  struct {
    TimerCallbackFunc func;
    TimerCallbackData data;
  } callback;
  struct Timer *next;
  struct TimerList *tl;
} Timer;

typedef struct TimerList {
  Timer *head;
} TimerList;

TimerFireAction Timer_fire(Timer *timer);

Timer *new_timer(TimerList *tl, TimerCallbackFunc func, TimerCallbackData data, size_t seconds);
void delete_timer(Timer *timer, bool delete_data);

void Timer_arm(Timer *timer);
void Timer_disarm(Timer *timer);
void Timer_rearm(Timer *timer);

void TimerList_examine(TimerList *tl);

#endif // !TIMER_H
