#ifndef NETWORK
#define NETWORK

#include <poll.h>

#define BUFSIZE 1000
#define BACKLOG 10

int start_listening(char *port);
void process_connections(int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds);
void handle_new_connection(int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds);

#endif // !NETWORK
