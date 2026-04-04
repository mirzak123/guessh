#include "timer.h"
#include "util.h"
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static void Timer_set(Timer *timer);

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
  timer->is_armed = false;
  timer->callback.func = func;
  timer->callback.data = data;
  timer->next = NULL;
  timer->tl = tl;

  return timer;
}

void delete_timer(Timer *timer, bool free_data) {
  if (timer->is_armed) {
    Timer_disarm(timer);
  }
  if (free_data) {
    free(timer->callback.data);
  }
  free(timer);
}

void Timer_fire(Timer *timer) {
  printf("Firing timer [%d]...\n", timer->id);
  timer->is_armed = false;
  if (timer->callback.func != NULL) {
    timer->callback.func(timer->callback.data);
  }
}

void TimerList_examine(TimerList *tl) {
  Timer *current = tl->head, *reschedule_list = NULL, *next = NULL;
  tl->head = NULL;
  while (current != NULL && current->timestamp <= time(NULL)) {
    next = current->next;
    Timer_fire(current);
    current = next;
  }

  reschedule_list = tl->head;
  tl->head = current;
  current = reschedule_list;

  while (current != NULL) {
    next = current->next;
    Timer_arm(current);
    current = next;
  }
}

void Timer_arm(Timer *timer) {
  TimerList *tl = timer->tl;

  Timer_set(timer);

  if (tl->head == NULL) {
    tl->head = timer;
    timer->next = NULL;
    return;
  }

  Timer *current = tl->head, *prev = NULL;

  while (current != NULL) {
    if (current == timer) {
      printf("[Timer_arm] Trying to add a timer that is already in the list\n");
      return;
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

  timer->is_armed = false;

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

  Timer_set(timer);

  Timer *current = timer->tl->head, *prev = NULL;
  while (current != timer) {
    prev = current;
    current = current->next;
  }

  // timer should stay in the same position
  if (timer->next != NULL && timer->timestamp <= timer->next->timestamp) {
    return;
  }

  // timer needs to shift
  if (prev == NULL) { // timer was head of the list
    timer->tl->head = timer->next;
  } else {
    prev->next = timer->next;
  }

  current = timer->next;
  while (current != NULL && current->timestamp <= timer->timestamp) {
    prev = current;
    current = current->next;
  }

  if (prev == NULL) { // only 1 timer in the list
    timer->tl->head = timer;
  } else {
    prev->next = timer;
    timer->next = current;
  }
}

void Timer_set(Timer *timer) {
  timer->is_armed = true;
  timer->timestamp = time(NULL) + timer->seconds;
}

void TimerList_print(TimerList *tl) {
  printf("[print_timer_list]: ");

  Timer *current = tl->head;
  while (current != NULL) {
    printf("[%d]", current->id);
    if (current->next != NULL)
      printf(" -> ");
    current = current->next;
  }
  printf("\n");
}
