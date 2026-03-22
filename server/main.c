#include <errno.h>
#include <poll.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <time.h>
#include <unistd.h>

#include "game_logic.h"
#include "game_server.h"
#include "network.h"
#include "timer.h"

#define ENV_PORT "GUESSH_SERVER_PORT"
#define DEFAULT_PORT "2480"
#define BUFF_LEN 1000
#define POLL_TIMEOUT 500

static void handle_shutdown(int sig);

volatile sig_atomic_t server_running = true;

int main(void) {
  GameServer *gs;
  int listen_fd;
  int fd_size = 5; // room for connections
  int fd_count;    // current connections
  struct pollfd *pfds = malloc(sizeof *pfds * fd_size);

  char *port = getenv(ENV_PORT);
  if (port == NULL) {
    port = DEFAULT_PORT;
  }

  listen_fd = start_listening(port);
  printf("Listening on port %s...\n", port);

  // Add listener to poll file descriptor list
  pfds[0].fd = listen_fd;
  pfds[0].events = POLLIN;
  fd_count = 1;

  srand(time(NULL));
  gs = GS_create();

  signal(SIGINT, handle_shutdown);
  signal(SIGTERM, handle_shutdown);

  while (server_running) {
    Timer_list_examine(&gs->timer_list);

    int poll_count = poll(pfds, fd_count, POLL_TIMEOUT);

    if (poll_count == -1) {
      if (errno == EINTR)
        // HACK: Stop poll crashing program while inserting a breakpoint when debugging
        continue;
      perror("poll");
      exit(1);
    }

    if (poll_count > 0) {
      process_connections(gs, listen_fd, &fd_size, &fd_count, &pfds);
    }
  }

  printf("\nShutting down gracefully...\n");
  GS_destroy(gs);
  free(pfds);

  return 0;
}

static void handle_shutdown(int sig) {
  (void)sig;
  server_running = false;
}
