#include "client_handler.h"
#include "game.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <time.h>

void handle_client(int client_fd) {
  srand(time(NULL));
  char client_msg[BUFSIZE], server_msg[BUFSIZE], *target_word;
  LetterFeedback feedback[WORD_LENGTH];
  int rv;

  target_word = get_random_word();
  printf("Target word: %s\n", target_word);

  strcpy(server_msg, "Hello there! Would you like to begin?\n");
  if ((rv = send(client_fd, server_msg, strlen(server_msg), 0)) == -1) {
    perror("send:");
    return;
  }

  if ((rv = recv(client_fd, client_msg, BUFSIZE, 0)) == -1) {
    perror("recv:");
    return;
  } else if (!rv) {
    return;
  }
  printf("rv: %d\n", rv);
  client_msg[rv] = '\0';

  strcpy(server_msg, "Guess...\n");
  if ((rv = send(client_fd, server_msg, strlen(server_msg), 0)) == -1) {
    perror("send:");
    return;
  }

  if ((rv = recv(client_fd, client_msg, BUFSIZE, 0)) == -1) {
    perror("recv:");
    return;
  } else if (!rv) {
    return;
  }
  client_msg[rv] = '\0';

  if (evaluate_guess(client_msg, target_word, feedback, WORD_LENGTH) == 1) {
    strcpy(server_msg, "You are correct!\n");
  } else {
    sprintf(server_msg, "Sorry, your word was: %s\n", target_word);
  }
  printf("msg: %s\n", server_msg);

  if ((rv = send(client_fd, server_msg, strlen(server_msg), 0)) == -1) {
    perror("send:");
    return;
  }
}
