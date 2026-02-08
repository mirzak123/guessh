#ifndef HASH_TABLE_H
#define HASH_TABLE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#define KEY(x)                                                                                                                   \
  _Generic((x),                                                                                                                  \
      _Bool: _ht_internal_key_int,                                                                                               \
      int: _ht_internal_key_int,                                                                                                 \
      long long: _ht_internal_key_int,                                                                                           \
      char *: _ht_internal_key_str,                                                                                              \
      const char *: _ht_internal_key_str)(x)

typedef enum { HT_KEY_EMPTY, HT_KEY_TOMBSTONE, HT_KEY_INT, HT_KEY_STR } KeyType;

typedef struct {
  KeyType type;
  union {
    int64_t i64;
    const char *str;
  };
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

static inline Key _ht_internal_key_int(int64_t v) { return (Key){.type = HT_KEY_INT, .i64 = v}; }
static inline Key _ht_internal_key_str(const char *s) { return (Key){.type = HT_KEY_STR, .str = s}; }

#endif
