#ifndef NETWORK
#define NETWORK

#include "game_server.h"
#include <poll.h>
#include <stdint.h>

int start_listening(char *port);
void process_connections(GameServer *gs, int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds);
void handle_new_connection(GameServer *gs, int listen_fd, int *fd_size, int *fd_count, struct pollfd **pfds);

#endif // !NETWORK
