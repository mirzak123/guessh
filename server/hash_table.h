#ifndef HASH_TABLE_H
#define HASH_TABLE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#define KEY_INT(x)                                                                                                               \
  (Key) { (uint8_t *)&x, sizeof(int) }

#define KEY_STR(x)                                                                                                               \
  (Key) { (uint8_t *)x, strlen(x) }

typedef struct {
  const uint8_t *data;
  size_t size;
} Key;

typedef const void *Value;

typedef struct {
  Key key;
  Value value;
} Entry;

typedef struct {
  int count;
  int capacity;
  Entry *entries;
} HashTable;

HashTable *HT_create(void);
void HT_destroy(HashTable *table);
void HT_set(HashTable *table, Key key, Value value);
Value HT_get(HashTable *table, Key key);
void HT_delete(HashTable *table, Key key);

#endif
