#include <stdio.h>
#include <unistd.h>

void sleep_with_log(unsigned int seconds) {
  printf("sleeping for %d seconds...\n", seconds);
  sleep(seconds);
}
