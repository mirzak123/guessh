#ifndef TIMER_H
#define TIMER_H

#include <stddef.h>

typedef struct Timer {
  size_t timestamp;
  struct Timer *next;
} Timer;

Timer *new_timer(size_t seconds);
void delete_timer(Timer *timer);

void check_timers(Timer *head);
void add_timer(Timer **head, Timer *timer);

#endif // !TIMER_H
