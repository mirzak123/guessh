#ifndef JSON_MESSAGES_H
#define JSON_MESSAGES_H

#include "game_logic.h"
#include <cjson/cJSON.h>
#include <stdbool.h>

cJSON *json_error(const char *reason);
cJSON *json_room_created(const char *room_id);
cJSON *json_match_started(const char *match_id, int rounds, size_t word_len);
cJSON *json_round_started(int round_num);
cJSON *json_guess_result(bool success, const char *guess, const LetterFeedback *feedback, size_t word_len);
cJSON *json_round_finished(bool success, const char *word);
cJSON *json_match_finished(const char *winner);

#endif
