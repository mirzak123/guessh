#include "timer.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

Timer *new_timer(int seconds) {
  Timer *timer = malloc(sizeof(Timer));
  if (timer == NULL) {
    perror("malloc");
    return NULL;
  }

  time_t current_time = time(NULL);
  if (current_time == -1) {
    perror("time");
    free(timer);
    return NULL;
  }

  timer->timestamp = current_time + seconds;
  timer->next = NULL;

  return timer;
}

void delete_timer(Timer *timer) { free(timer); }

void add_timer(Timer **head, Timer *timer) {
  if (*head == NULL) {
    *head = timer;
    return;
  }

  Timer *current = *head, *prev = NULL;

  while (current != NULL) {
    if (current->timestamp > timer->timestamp) {
      timer->next = current;

      if (prev == NULL) { // inserting at the beginning of the list
        *head = timer;
        return;
      }

      prev->next = timer;
      timer->next = current;
      return;
    }

    prev = current;
    current = current->next;
  }

  current->next = timer;
}
