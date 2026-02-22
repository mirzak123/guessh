#ifndef JSON_MESSAGES_H
#define JSON_MESSAGES_H

#include "game_logic.h"
#include "game_types.h"
#include <cjson/cJSON.h>
#include <stdbool.h>

cJSON *json_error(const char *reason);
cJSON *json_room_created(const char *room_id);
cJSON *json_room_joined(const char *room_id);
cJSON *json_room_join_failed(const char *room_id, const char *reason);
cJSON *json_match_started(const char *match_id, int rounds, size_t word_len, char *opponent_name);
cJSON *json_round_started(size_t round_num, size_t max_attempts);
cJSON *json_guess_result(bool success, const char *guess, const LetterFeedback *feedback, size_t word_len);
cJSON *json_round_finished(Outcome outcome, const char *word);
cJSON *json_match_finished(Outcome outcome);
cJSON *json_opponent_typing(const char *value);

#endif
