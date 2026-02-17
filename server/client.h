#ifndef CLIENT_H
#define CLIENT_H

#include "game_types.h"
#include <cjson/cJSON.h>
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
  Player *player;
} Client;

Client *new_client(int client_fd);
void delete_client(Client *client);

void send_json(int client_fd, cJSON *json);
void send_only_type(int client_fd, const char *type);
void send_error(int client_fd, const char *reason);

#endif // !CLIENT_H
