#include "hash_table.h"
#include <_stdio.h>
#include <_string.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define INITIAL_CAPACITY 8
#define TABLE_MAX_LOAD 0.75
#define GROW_CAPACITY(capacity) ((capacity) < INITIAL_CAPACITY ? INITIAL_CAPACITY : (capacity) * 2)

static Entry *find_entry(Entry *entries, int capacity, Key *key);
static void adjust_capacity(HashTable *table, int capacity);
static uint32_t hash_key(Key *key);
static bool compare_keys(Key *k1, Key *k2);
static inline bool is_empty(Entry *entry);
static inline bool is_tombstone(Entry *entry);

HashTable *HT_create(void) {
  HashTable *table = malloc(sizeof(HashTable));

  if (!table) {
    perror("malloc");
    exit(1);
  }

  table->capacity = 0;
  table->count = 0;
  table->entries = NULL;
  return table;
}

void HT_destroy(HashTable *table) {
  for (int i = 0; i < table->capacity; i++)
    if (table->entries[i].key.type == HT_KEY_STR)
      free(table->entries[i].key.str);

  free(table->entries);
  free(table);
}

void HT_set(HashTable *table, Key key, Value value) {
  if (table->count + 1 > table->capacity * TABLE_MAX_LOAD) {
    int capacity = GROW_CAPACITY(table->capacity);
    adjust_capacity(table, capacity);
  }

  Entry *entry = find_entry(table->entries, table->capacity, &key);
  if (is_empty(entry)) {
    table->count++;
  } else {
    if (entry->key.type == HT_KEY_STR) {
      free(entry->key.str);
    }
  }

  if (key.type == HT_KEY_STR) {
    key.str = strdup(key.str);
  }
  entry->key = key;
  entry->value = value;
}

Value HT_get(HashTable *table, Key key) {
  if (table->count == 0)
    return NULL;
  Entry *entry = find_entry(table->entries, table->capacity, &key);
  return entry == NULL ? NULL : entry->value;
}

void HT_delete(HashTable *table, Key key) {
  if (table->count == 0)
    return;

  Entry *entry = find_entry(table->entries, table->capacity, &key);
  if (!is_empty(entry)) {
    if (entry->key.type == HT_KEY_STR) {
      free(entry->key.str);
    }
    entry->key.type = HT_KEY_TOMBSTONE;
    entry->value = NULL;
  }
}

static Entry *find_entry(Entry *entries, int capacity, Key *key) {
  int index = hash_key(key) % capacity;
  Entry *tombstone = NULL;

  while (true) {
    Entry *entry = &entries[index];
    if (is_empty(entry)) {
      return tombstone != NULL ? tombstone : entry;
    }

    if (compare_keys(key, &entry->key)) {
      return entry;
    }

    if (is_tombstone(entry) && tombstone == NULL) {
      tombstone = entry;
    }

    index = (index + 1) % capacity;
  }
}

static void adjust_capacity(HashTable *table, int capacity) {
  Entry *entries = malloc(sizeof(Entry) * capacity);
  for (int i = 0; i < capacity; i++) {
    Entry *entry = &entries[i];
    entry->key = (Key){.type = HT_KEY_EMPTY};
    entry->value = NULL;
  }

  table->count = 0;

  for (int i = 0; i < table->capacity; i++) {
    Entry *entry = &table->entries[i];
    if (is_empty(entry) || is_tombstone(entry))
      continue;

    Entry *new_entry = find_entry(entries, capacity, &entry->key);
    new_entry->key = entry->key;
    new_entry->value = entry->value;
    table->count++;
  }
  table->capacity = capacity;
  free(table->entries);
  table->entries = entries;
}

/* FNV-1a hashing algorithm */
static uint32_t hash_key(Key *key) {
  uint32_t hash = 2166136261u;
  uint8_t *data = NULL;
  int size = 0;

  switch (key->type) {
  case HT_KEY_INT:
    data = (uint8_t *)&key->i64;
    size = sizeof(key->i64);
    break;
  case HT_KEY_STR:
    data = (uint8_t *)key->str;
    size = strlen(key->str);
    break;
  case HT_KEY_EMPTY:
  case HT_KEY_TOMBSTONE:
    perror("trying to hash invalid key");
    exit(1);
  }

  for (int i = 0; i < size; i++) {
    hash ^= data[i];
    hash *= 16777619;
  }
  return hash;
}

static bool compare_keys(Key *k1, Key *k2) {
  if (k1->type != k2->type)
    return false;
  switch (k1->type) {
  case HT_KEY_INT:
    return k1->i64 == k2->i64;
  case HT_KEY_STR:
    return !strcmp(k1->str, k2->str);
  default:
    return true;
  }
}

static inline bool is_empty(Entry *entry) { return entry->key.type == HT_KEY_EMPTY; }
static inline bool is_tombstone(Entry *entry) { return entry->key.type == HT_KEY_TOMBSTONE; }
