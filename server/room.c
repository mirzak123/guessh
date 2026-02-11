#include "room.h"
#include "util.h"
#include <stdlib.h>

Room *new_room(void) {
  Room *room = malloc(sizeof(Room));
  room->id = generate_random_string(ROOM_ID_LEN);
  room->player1 = NULL;
  room->player2 = NULL;
  return room;
}

void delete_room(Room *room) {
  free(room->id);
  free(room);
}
