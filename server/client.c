#include "client.h"
#include "game_types.h"
#include "json_messages.h"

#include <arpa/inet.h>
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>

Client *new_client(int client_fd) {
  Client *client = malloc(sizeof(Client));
  if (client == NULL) {
    perror("malloc");
    return NULL;
  }

  client->fd = client_fd;
  client->buf_start = client->buffer;
  client->buf_len = 0;
  client->payload_size = 0;
  client->state = READING_LENGTH;
  client->player = NULL;

  return client;
}

void delete_client(Client *client) {
  assert(client != NULL);
  printf("Deleting client [fd: %d]\n", client->fd);

  if (client->player) {
    delete_player(client->player);
  }
  free(client);
}

void send_json(int client_fd, cJSON *json) {
  size_t length;
  uint32_t nlength;
  char *message = cJSON_PrintUnformatted(json);

  if (message == NULL) {
    printf("[send_json] cJSON_PrintUnformatted() failed");
    return;
  }

  // TCP segment length prefix
  length = strlen(message);
  nlength = htonl(length);
  printf("[send_json] sending %zu bytes of data [fd: %d]: %s\n", length, client_fd, message);
  if (send(client_fd, &nlength, 4, 0) == -1) {
    perror("send");
  }

  // Data
  if (send(client_fd, message, strlen(message), 0) == -1) {
    perror("send");
  }

  cJSON_free(message);
}

void send_only_type(int client_fd, const char *type) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", type);
  send_json(client_fd, json);
  cJSON_Delete(json);
}

void send_error(int client_fd, const char *reason) {
  cJSON *err_json = json_error(reason);
  send_json(client_fd, err_json);
  cJSON_Delete(err_json);
}
