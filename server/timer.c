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

void delete_timer(Timer *timer, bool delete_data) {
  if (delete_data) {
    free(timer->callback.data);
  }
  free(timer);
}

bool Timer_fire(Timer *timer) {
  printf("Firing timer [%d]...\n", timer->id);
  if (timer->callback.func != NULL) {
    return timer->callback.func(timer->callback.data);
  }
  return false;
}

void Timer_list_examine(Timer **head) {
  Timer *reset_list = NULL, *next = NULL;
  while (*head != NULL && (*head)->timestamp <= time(NULL)) {
    next = (*head)->next;
    bool should_reset = Timer_fire(*head);
    if (should_reset) {
      (*head)->next = reset_list;
      reset_list = *head;
    }
    *head = next;
  }

  while (reset_list != NULL) {
    next = reset_list->next;
    reset_list->timestamp = time(NULL) + reset_list->seconds;
    Timer_list_add(head, reset_list);
    reset_list = next;
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
