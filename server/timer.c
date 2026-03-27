#include "timer.h"
#include "util.h"
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

Timer *new_timer(size_t seconds, TimerCallbackFunc func, TimerCallbackData data) {
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
  timer->seconds = seconds;
  timer->callback.func = func;
  timer->callback.data = data;
  timer->next = NULL;

  return timer;
}

void delete_timer(Timer *timer, bool free_data) {
  if (free_data) {
    free(timer->callback.data);
  }
  free(timer);
}

TimerFireAction Timer_fire(Timer *timer) {
  printf("Firing timer [%d]...\n", timer->id);
  if (timer->callback.func != NULL) {
    return timer->callback.func(timer->callback.data);
  }
  return TIMER_FIRE_NONE;
}

void TimerList_examine(TimerList *tl) {
  Timer *rearm_list = NULL, *next = NULL;
  while (tl->head != NULL && tl->head->timestamp <= time(NULL)) {
    next = tl->head->next;
    if (Timer_fire(tl->head) == TIMER_FIRE_REARM) {
      tl->head->next = rearm_list;
      rearm_list = tl->head;
    }
    tl->head = next;
  }

  while (rearm_list != NULL) {
    next = rearm_list->next;
    rearm_list->timestamp = time(NULL) + rearm_list->seconds;
    TimerList_add(tl, rearm_list);
    rearm_list = next;
  }
}

void TimerList_add(TimerList *tl, Timer *timer) {
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

void TimerList_remove(TimerList *tl, Timer *timer) {
  Timer *current = tl->head, *prev = NULL;

  while (current != NULL) {
    if (current != timer) {
      prev = current;
      current = current->next;
      continue;
    }

    if (prev == NULL) {
      tl->head = current->next;
    } else {
      prev->next = current->next;
    }
    return;
  }
}

void TimerList_rearm(TimerList *tl, Timer *timer) {
  if (timer == NULL) {
    printf("Trying to reset a NULL timer\n");
    return;
  }

  TimerList_remove(tl, timer);
  timer->timestamp = time(NULL) + timer->seconds;
  timer->next = NULL;
  TimerList_add(tl, timer);
}
