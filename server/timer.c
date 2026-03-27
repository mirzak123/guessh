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

void Timer_list_examine(Timer **head) {
  Timer *rearm_list = NULL, *next = NULL;
  while (*head != NULL && (*head)->timestamp <= time(NULL)) {
    next = (*head)->next;
    if (Timer_fire(*head) == TIMER_FIRE_REARM) {
      (*head)->next = rearm_list;
      rearm_list = *head;
    }
    *head = next;
  }

  while (rearm_list != NULL) {
    next = rearm_list->next;
    rearm_list->timestamp = time(NULL) + rearm_list->seconds;
    Timer_list_add(head, rearm_list);
    rearm_list = next;
  }
}

void Timer_list_add(Timer **head, Timer *timer) {
  if (*head == NULL) {
    *head = timer;
    return;
  }

  Timer *current = *head, *prev = NULL;

  while (current != NULL) {
    if (current == timer) {
      return; // trying to add a timer that is already in the list
    }

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

void Timer_list_rearm(Timer **head, Timer *timer) {
  if (timer == NULL) {
    printf("Trying to reset a NULL timer\n");
    return;
  }

  Timer_list_remove(head, timer);
  timer->timestamp = time(NULL) + timer->seconds;
  timer->next = NULL;
  Timer_list_add(head, timer);
}
