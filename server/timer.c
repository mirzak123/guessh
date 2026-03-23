#include "timer.h"
#include "util.h"
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

void delete_timer(Timer *timer) { free(timer); }

void Timer_fire(Timer *timer) {
  printf("Firing timer [%d]...\n", timer->id);
  if (timer->callback.func != NULL) {
    timer->callback.func(timer->callback.data);
  }
}

void Timer_list_examine(Timer **head) {
  Timer *current = *head, *next;

  /* NOTE: To future self:
   * If Timer_fire want to reset the timer, and we insert the timer into the same list
   * we are traversing, it will skip it with `*head = (*head)->next`. This is why we
   * point the head to an empty list. Any timer that wants to reset will be added to this
   * empty list within Timer_fire, and any timer that did not fire in this round, we add
   * back to the list.
   *
   * TODO: This defeats the purpose of the linked list timing scheme, as we need to loop through
   * all timers in every iteration, so I should really think of a solution to reset the timers on the
   * original list without breaking stuff. Works for now. */
  *head = NULL;
  while (current != NULL && (current)->timestamp <= time(NULL)) {
    next = current->next;
    Timer_fire(current);
    current = next;
    printf("first loop\n");
  }

  while (current != NULL) {
    Timer_list_add(head, current);
    current = current->next;
    printf("second loop\n");
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

void Timer_list_reset(Timer **head, Timer *timer) {
  if (timer == NULL) {
    printf("Trying to reset a NULL timer\n");
    return;
  }

  Timer_list_remove(head, timer);
  timer->timestamp = time(NULL) + timer->seconds;
  timer->next = NULL;
  Timer_list_add(head, timer);
}
