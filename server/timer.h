#ifndef TIMER_H
#define TIMER_H

#include <stdbool.h>
#include <stddef.h>

typedef bool (*TimerCallbackFunc)(void *data);
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
void delete_timer(Timer *timer);

void Timer_list_examine(Timer **head);
void Timer_list_add(Timer **head, Timer *timer);
void Timer_list_remove(Timer **head, Timer *timer);
void Timer_list_reset(Timer **head, Timer *timer);

#endif // !TIMER_H
