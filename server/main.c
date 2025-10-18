#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <unistd.h>

#include "client_handler.h"
#include "network.h"

#define PORT "2480"
#define BUFF_LEN 1000

int main() {
  int listen_fd, client_fd;

  printf("PORT: %s\n", PORT);

  listen_fd = start_listening(PORT);

  while (1) {
    client_fd = get_client_conn(listen_fd);

    if (!fork()) {
      close(listen_fd);

      handle_client(client_fd);

      close(client_fd);
      exit(0);
    }
    close(client_fd);
  }

  return 0;
}
