#ifndef ROOM_H
#define ROOM_H

#include "game_types.h"
#define ROOM_ID_LEN 5

typedef struct {
  char *id;
  Player *player1;
  Player *player2;
} Room;

Room *new_room(void);
void delete_room(Room *room);

#endif // !ROOM_H
