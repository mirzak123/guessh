#include "network.h"
#include "client.h"
#include "game_server.h"
#include "game_types.h"
#include <arpa/inet.h>
#include <netdb.h>
#include <netinet/in.h>
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/errno.h>
#include <sys/poll.h>
#include <sys/signal.h>
#include <sys/socket.h>
#include <unistd.h>

void sigchld_handler(void) {
  // waitpid() might overwrite errno, so we save and restore it:
  int saved_errno = errno;

  while (waitpid(-1, NULL, WNOHANG) > 0)
    ;

  errno = saved_errno;
}

int start_listening(char *port) {
  int status, sock_fd;
  struct addrinfo hints, *servinfo, *p;
  struct sigaction sa;
  int yes = 1;

  memset(&hints, 0, sizeof hints);
  hints.ai_family = AF_UNSPEC;
  hints.ai_socktype = SOCK_STREAM;
  hints.ai_flags = AI_PASSIVE;

  if ((status = getaddrinfo("localhost", port, &hints, &servinfo)) != 0) {
    fprintf(stderr, "gai error: %s\n", gai_strerror(status));
    exit(1);
  }

  for (p = servinfo; p != NULL; p = p->ai_next) {
    void *addr;
    char ipstr[INET_ADDRSTRLEN];
    struct sockaddr_in *ipv4;

    if (p->ai_family == AF_INET) {
      ipv4 = (struct sockaddr_in *)p->ai_addr;
      addr = &(ipv4->sin_addr);
      inet_ntop(p->ai_family, addr, ipstr, sizeof ipstr);
      printf("Using address: %s\n", ipstr);
      break;
    }
  }

  if ((sock_fd = socket(p->ai_family, p->ai_socktype, p->ai_protocol)) == -1) {
    perror("socket");
    exit(errno);
  }

  // Prevents "Address already in use" error
  setsockopt(sock_fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof yes);

  if (bind(sock_fd, p->ai_addr, p->ai_addrlen) == -1) {
    perror("bind");
    exit(errno);
  }

  if (listen(sock_fd, BACKLOG) == -1) {
    perror("listen");
    exit(errno);
  }

  // NOTE: No clue how this code block works
  // ???
  sa.sa_handler = (void (*)(int))sigchld_handler; // reap all dead processes
  sigemptyset(&sa.sa_mask);
  sa.sa_flags = SA_RESTART;
  if (sigaction(SIGCHLD, &sa, NULL) == -1) {
    perror("sigaction");
    exit(1);
  }
  // ???

  freeaddrinfo(servinfo);

  return sock_fd;
}

void add_to_pfds(struct pollfd **pfds, int new_fd, int *fd_size, int *fd_count) {
  if (*fd_size == *fd_count) {
    *fd_size *= 2;
    *pfds = realloc(*pfds, sizeof(**pfds) * (*fd_size));
  }

  (*pfds)[*fd_count].fd = new_fd;
  (*pfds)[*fd_count].events = POLLIN;
  (*pfds)[*fd_count].revents = 0;

  (*fd_count)++;
}

/*
 * Delete a file descriptor after connection closes.
 */
void del_from_pfds(struct pollfd pfds[], int i, int *fd_count) {
  pfds[i] = pfds[*fd_count - 1];
  (*fd_count)--;
}

/*
 * Handle incoming connections.
 */
void handle_new_connection(GameServer *gs, int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds) {
  struct sockaddr_storage client_addr;
  socklen_t addr_size = sizeof client_addr;

  int client_fd = accept(listen_fd, (struct sockaddr *)&client_addr, &addr_size);

  if (client_fd == -1) {
    perror("accept");
    return;
  }

  add_to_pfds(pfds, client_fd, fd_size, fd_count);

  gs->clients[client_fd] = new_client(client_fd);
}

/*
 * Handle client data or client hangup.
 */
void handle_client_data(GameServer *gs, int *fd_count, struct pollfd pfds[], int *pfd_i) {
  char incoming_buf[BUFSIZE];

  int client_fd = pfds[*pfd_i].fd;
  Client *client = gs->clients[client_fd];

  int nbytes = recv(client_fd, incoming_buf, BUFSIZE, 0);

  if (nbytes <= 0) {
    if (nbytes == 0) { // client hang up
      printf("socket %d hangup\n", client_fd);
    } else { // error
      perror("recv");
    }

    close(client_fd);
    del_from_pfds(pfds, *pfd_i, fd_count);

    // Delete match if it exists
    Match *match = GS_get_match_by_client_fd(gs, client_fd);
    if (match != NULL) {
      // TODO: Handle premature match end for multiplayer games better by notifying
      // the other client correctly on why the match ended
      GS_end_match(gs, match);
    }

    // re-examine slot as it contains a new fd after deletion
    (*pfd_i)--;

    return;
  }

  printf("[handle_client_data] received %d bytes of data from fd %d\n", nbytes, client_fd);

  if ((client->buf_len + nbytes) - (client->buf_start - client->buffer) > BUFSIZE) {
    printf("[handle_client_data] error: buffer overflow\n");
    close(client_fd);
    del_from_pfds(pfds, *pfd_i, fd_count);

    Match *match = GS_get_match_by_client_fd(gs, client_fd);
    if (match != NULL) {
      GS_end_match(gs, match);
    }
    (*pfd_i)--;
    return;
  }

  memcpy(client->buffer + client->buf_len, incoming_buf, nbytes);
  client->buf_len += nbytes;

  int run = 1;
  while (run) {
    switch (client->state) {
    case READING_LENGTH:
      if (client->buf_len < LEN_PREFIX_BYTES) {
        run = 0;
        break;
      }

      uint32_t netlen;
      memcpy(&netlen, client->buf_start, LEN_PREFIX_BYTES);

      client->payload_size = ntohl(netlen);
      if (client->payload_size > BUFSIZE) {
        printf("[handle_client_data] error: payload size %d larger than allowed buffer limit %d\n", client->payload_size,
               BUFSIZE);
        close(client_fd);
        return;
      }

      client->buf_start += LEN_PREFIX_BYTES;
      client->buf_len -= LEN_PREFIX_BYTES;
      client->state = READING_PAYLOAD;
      break;

    case READING_PAYLOAD:
      if (client->buf_len < client->payload_size) {
        run = 0;
        break;
      }

      GS_handle_request(gs, client_fd, client->buf_start, client->payload_size);

      client->buf_start += client->payload_size;
      client->buf_len -= client->payload_size;
      client->state = READING_LENGTH;
      break;
    }
  }

  memmove(client->buffer, client->buf_start, client->buf_len);
  client->buf_start = client->buffer;
}

void process_connections(GameServer *gs, int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds) {
  for (int i = 0; i < *fd_count; i++) {
    if ((*pfds)[i].revents & (POLLIN | POLLHUP)) {
      if ((*pfds)[i].fd == listen_fd) {
        handle_new_connection(gs, listen_fd, fd_size, fd_count, pfds);
      } else {
        handle_client_data(gs, fd_count, *pfds, &i);
      }
    }
  }
}
