#include "json_messages.h"
#include "game_logic.h"
#include "game_server.h"
#include "game_types.h"
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

cJSON *json_room_joined(const char *room_id) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(ROOM_JOINED));
  cJSON_AddStringToObject(json, "roomId", room_id);
  return json;
}

cJSON *json_room_join_failed(const char *room_id, const char *reason) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(ROOM_JOIN_FAILED));
  cJSON_AddStringToObject(json, "roomId", room_id);
  cJSON_AddStringToObject(json, "reason", reason);
  return json;
}

cJSON *json_match_started(const char *match_id, int rounds, size_t word_len, char *opponent_name) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(MATCH_STARTED));
  cJSON_AddStringToObject(json, "matchId", match_id);
  cJSON_AddNumberToObject(json, "rounds", rounds);
  cJSON_AddNumberToObject(json, "wordLength", word_len);
  if (opponent_name != NULL) {
    cJSON_AddStringToObject(json, "opponentName", opponent_name);
  }
  return json;
}

cJSON *json_round_started(size_t round_num, size_t max_attempts) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(ROUND_STARTED));
  cJSON_AddNumberToObject(json, "roundNumber", round_num);
  cJSON_AddNumberToObject(json, "maxAttempts", max_attempts);
  return json;
}

cJSON *json_guess_result(bool success, const char *guess, const LetterFeedback *feedback, size_t word_len) {
  cJSON *json = cJSON_CreateObject(), *feedback_json = cJSON_CreateArray(), *feedback_item = NULL;

  cJSON_AddStringToObject(json, "type", STR(GUESS_RESULT));
  cJSON_AddBoolToObject(json, "success", success);
  cJSON_AddStringToObject(json, "guess", guess);
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

cJSON *json_match_finished(Outcome outcome) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(MATCH_FINISHED));
  cJSON_AddNumberToObject(json, "outcome", outcome);
  return json;
}

cJSON *json_opponent_typing(const char *value) {
  cJSON *json = cJSON_CreateObject();
  cJSON_AddStringToObject(json, "type", STR(OPPONENT_TYPING));
  cJSON_AddStringToObject(json, "value", value);
  return json;
}
