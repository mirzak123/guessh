#ifndef ROOM_H
#define ROOM_H

#define ROOM_KEY_LEN 5

typedef struct {
  char *room_id;
  int player_num;
  struct Room *next;
} Room;

Room *Room_create(void);
void Room_destroy(Room *room);

#endif // !ROOM_H
