#ifndef CLIENT_H
#define CLIENT_H

#include <stddef.h>
#include <stdint.h>

#define BUFSIZE 1000

typedef enum {
  READING_LENGTH,
  READING_PAYLOAD,
} ClientState;

typedef struct {
  int fd;
  ClientState state;
  char buffer[BUFSIZE];
  char *buf_start;
  size_t buf_len;
  uint32_t payload_size;
} Client;

Client *new_client(int client_fd);
void delete_client(Client *client);

#endif // !CLIENT_H
