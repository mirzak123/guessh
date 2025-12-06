#include "client.h"
#include <stdio.h>
#include <stdlib.h>

Client *new_client(int client_fd) {
  Client *client = malloc(sizeof(Client));
  if (client == NULL) {
    perror("malloc");
    return NULL;
  }

  client->fd = client_fd;
  client->buffer_len = 0;
  client->payload_size = 0;
  client->state = READING_LENGTH;

  return client;
}

void delete_client(Client *client) { free(client); }
