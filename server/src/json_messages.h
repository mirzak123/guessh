#ifndef JSON_MESSAGES_H
#define JSON_MESSAGES_H

#include "game_server.h"
#include "game_types.h"
#include <cjson/cJSON.h>
#include <stdbool.h>

cJSON *json_error(const char *reason);
cJSON *json_room_created(const char *room_id);
cJSON *json_room_joined(const char *room_id);
cJSON *json_room_join_failed(const char *room_id, const char *reason);
cJSON *json_match_started(const char *match_id, GameFormat format, int rounds, size_t word_len, Timer *turn_timer,
                          char *opponent_name);
cJSON *json_round_started(size_t round_num, size_t max_attempts);
cJSON *json_guess_result(const char *guess, Round *round, size_t word_len);
cJSON *json_round_finished(int points, const char **words, int len, Timer *post_round_timer);
cJSON *json_match_finished(Outcome outcome, bool opponentLeft);
cJSON *json_opponent_typing(const char *value);
cJSON *json_stats(ServerStats *stats);

#endif
