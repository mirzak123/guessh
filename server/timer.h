#ifndef TIMER_H
#define TIMER_H

#include <stddef.h>

typedef struct Timer {
  int id;
  size_t timestamp;
  struct Timer *next;
} Timer;

Timer *new_timer(size_t seconds);
void delete_timer(Timer *timer);

void Timer_list_check(Timer **head);
void Timer_list_add(Timer **head, Timer *timer);
void Timer_list_remove(Timer **head, Timer *timer);

#endif // !TIMER_H
