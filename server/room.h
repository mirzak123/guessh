#ifndef ROOM_H
#define ROOM_H

#define ROOM_ID_LEN 5

struct Player;
struct Match;

typedef struct Room {
  char *id;
  struct Player *player1;
  struct Player *player2;
  struct Match *match;
} Room;

Room *new_room(void);
void delete_room(Room *room);

#endif // !ROOM_H
