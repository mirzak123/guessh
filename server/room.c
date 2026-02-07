#include "room.h"
#include "util.h"
#include <stdlib.h>

Room *Room_create(void) {
  Room *room = malloc(sizeof(Room));
  room->room_id = generate_random_string(ROOM_KEY_LEN);
  return room;
}

void Room_destroy(Room *room) {
  free(room->room_id);
  free(room);
}
