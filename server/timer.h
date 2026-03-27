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
} Timer;

Timer *new_timer(size_t seconds, TimerCallbackFunc func, TimerCallbackData data);
void delete_timer(Timer *timer, bool delete_data);

TimerFireAction Timer_fire(Timer *timer);

typedef struct TimerList {
  Timer *head;
} TimerList;

void TimerList_examine(TimerList *tl);
void TimerList_add(TimerList *tl, Timer *timer);
void TimerList_remove(TimerList *tl, Timer *timer);
void TimerList_rearm(TimerList *tl, Timer *timer);

#endif // !TIMER_H
