#include <arpa/inet.h>
#include <netdb.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/_types/_socklen_t.h>
#include <sys/errno.h>
#include <sys/socket.h>

#define BACKLOG 10

int start_listening(char *port) {
  int status, sock_fd;
  struct addrinfo hints, *servinfo, *p;

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
    perror("socket:");
    exit(errno);
  }

  // Prevents "Address already in use" error
  int yes = 1;
  setsockopt(sock_fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof yes);

  if (bind(sock_fd, p->ai_addr, p->ai_addrlen) == -1) {
    perror("bind:");
    exit(errno);
  }

  if (listen(sock_fd, BACKLOG) == -1) {
    perror("listen:");
    exit(errno);
  }

  freeaddrinfo(servinfo);

  return sock_fd;
}

int get_client_conn(int fd) {
  struct sockaddr_storage client_addr;
  socklen_t addr_size = sizeof client_addr;

  int client_fd = accept(fd, (struct sockaddr *)&client_addr, &addr_size);
  return client_fd;
}

void sigchld_handler(int s) {
  // waitpid() might overwrite errno, so we save and restore it:
  int saved_errno = errno;

  while (waitpid(-1, NULL, WNOHANG) > 0)
    ;

  errno = saved_errno;
}
