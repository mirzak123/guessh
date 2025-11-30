#include "game_logic.h"
#include "game_server.h"
#include <cjson/cJSON.h>
#include <stdbool.h>

cJSON *json_error(const char *reason) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(ERROR));
  cJSON_AddStringToObject(json, "reason", reason);
  return json;
}

cJSON *json_room_created(const char *room_id) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(ROOM_CREATED));
  cJSON_AddStringToObject(json, "roomId", room_id);
  return json;
}

cJSON *json_match_started(const char *match_id, int rounds, size_t word_len) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(MATCH_STARTED));
  cJSON_AddStringToObject(json, "matchId", match_id);
  cJSON_AddNumberToObject(json, "rounds", rounds);
  cJSON_AddNumberToObject(json, "wordLength", word_len);
  return json;
}

cJSON *json_round_started(int round_num) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(ROUND_STARTED));
  cJSON_AddNumberToObject(json, "roundNumber", round_num);
  return json;
}

cJSON *json_guess_result(bool success, const LetterFeedback *feedback, size_t word_len) {
  cJSON *json = cJSON_CreateObject(), *feedback_json = cJSON_CreateArray(), *feedback_item = NULL;

  cJSON_AddStringToObject(json, "type", STR(GUESS_RESULT));
  cJSON_AddBoolToObject(json, "success", success);
  for (int i = 0; i < (int)word_len; i++) {
    feedback_item = cJSON_CreateNumber(feedback[i]);
    cJSON_AddItemToArray(feedback_json, feedback_item);
  }
  cJSON_AddItemToObject(json, "feedback", feedback_json);

  return json;
}

cJSON *json_round_finished(bool success, const char *word) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(ROUND_FINISHED));
  cJSON_AddBoolToObject(json, "success", success);
  cJSON_AddStringToObject(json, "word", word);
  return json;
}

cJSON *json_match_finished(const char *winner) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(MATCH_FINISHED));
  cJSON_AddStringToObject(json, "winner", winner);
  return json;
}
