#include "timer.h"
#include "util.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

Timer *new_timer(size_t seconds, CallbackFunc func, CallbackData data) {
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

  timer->id = generate_unique_id();
  timer->timestamp = current_time + seconds;
  timer->callback.func = func;
  timer->callback.data = data;
  timer->next = NULL;

  return timer;
}

void delete_timer(Timer *timer) { free(timer); }

void Timer_fire(Timer *timer) {
  if (timer->callback.func != NULL) {
    timer->callback.func(timer->callback.data);
  }
}

void Timer_list_add(Timer **head, Timer *timer) {
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

  prev->next = timer; // inserting at the end of the list
}

void Timer_list_remove(Timer **head, Timer *timer) {
  Timer *current = *head, *prev = NULL;

  while (current != NULL) {
    if (current != timer) {
      prev = current;
      current = current->next;
      continue;
    }

    if (prev == NULL) {
      *head = current->next;
    } else {
      prev->next = current->next;
    }
    return;
  }
}
