#ifndef NETWORK
#define NETWORK

#include "game_server.h"
#include <poll.h>
#include <stdint.h>

#define BACKLOG 10
#define LEN_PREFIX_BYTES 4

int start_listening(char *port);
void process_connections(GameServer *gs, int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds);
void handle_new_connection(GameServer *gs, int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds);
void remove_client(GameServer *gs, Client *client, struct pollfd pfds[], int *pfd_i, int *fd_count);

#endif // !NETWORK
