#include "timer.h"
#include "util.h"
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

Timer *new_timer(TimerList *tl, TimerCallbackFunc func, TimerCallbackData data, size_t seconds) {
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
  timer->seconds = seconds;
  timer->callback.func = func;
  timer->callback.data = data;
  timer->next = NULL;
  timer->tl = tl;

  return timer;
}

void delete_timer(Timer *timer, bool free_data) {
  if (free_data) {
    free(timer->callback.data);
  }
  free(timer);
}

void Timer_fire(Timer *timer) {
  printf("Firing timer [%d]...\n", timer->id);
  if (timer->callback.func != NULL) {
    timer->callback.func(timer->callback.data);
  }
}

void TimerList_examine(TimerList *tl) {
  Timer *current = tl->head, *next = NULL;
  tl->head = NULL;
  while (current != NULL && current->timestamp <= time(NULL)) {
    next = current->next;
    Timer_fire(current);
    current = next;
  }

  while (current != NULL) {
    next = current->next;
    Timer_arm(current);
    current = next;
  }
}

void Timer_arm(Timer *timer) {
  TimerList *tl = timer->tl;
  if (tl->head == NULL) {
    tl->head = timer;
    return;
  }

  Timer *current = tl->head, *prev = NULL;

  while (current != NULL) {
    if (current == timer) {
      return; // trying to add a timer that is already in the list
    }

    if (current->timestamp > timer->timestamp) {
      timer->next = current;

      if (prev == NULL) { // inserting at the beginning of the list
        tl->head = timer;
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

void Timer_disarm(Timer *timer) {
  Timer *current = timer->tl->head, *prev = NULL;

  while (current != NULL) {
    if (current != timer) {
      prev = current;
      current = current->next;
      continue;
    }

    if (prev == NULL) {
      timer->tl->head = current->next;
    } else {
      prev->next = current->next;
    }
    return;
  }
}

void Timer_rearm(Timer *timer) {
  if (timer == NULL) {
    printf("Trying to rearm a NULL timer\n");
    return;
  }

  timer->timestamp = time(NULL) + timer->seconds;
  timer->next = NULL;
  Timer_arm(timer);
}
