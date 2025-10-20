#include <errno.h>
#include <poll.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/poll.h>
#include <sys/socket.h>
#include <unistd.h>

#include "game_server.h"
#include "network.h"

#define PORT "2480"
#define BUFF_LEN 1000

int main() {
  GameServer *gs;
  int listen_fd, client_fd;
  int fd_size = 5; // room for connections
  int fd_count;    // current connections
  struct pollfd *pfds = malloc(sizeof *pfds * fd_size);

  listen_fd = start_listening(PORT);
  printf("Listening on port %s...\n", PORT);

  // Add listener to poll file descriptor list
  pfds[0].fd = listen_fd;
  pfds[0].events = POLLIN;
  fd_count = 1;

  gs = GS_create();
  for (;;) {
    int poll_count = poll(pfds, fd_count, -1);

    if (poll_count == -1) {
      perror("poll");
      exit(1);
    }

    process_connections(gs, listen_fd, &fd_size, &fd_count, &pfds);
  }

  GS_destroy(gs);
  free(pfds);

  return 0;
}
