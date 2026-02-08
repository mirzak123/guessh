#include "hash_table.h"
#include <_stdio.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define INITIAL_CAPACITY 8
#define TABLE_MAX_LOAD 0.75
#define GROW_CAPACITY(capacity) ((capacity) < INITIAL_CAPACITY ? INITIAL_CAPACITY : (capacity) * 2)
#define TOMBSTONE_VALUE (uintptr_t *)1

static Entry *find_entry(Entry *entries, int capacity, Key *key);
static void adjust_capacity(HashTable *table, int capacity);
static uint32_t hash_key(Key k);
static inline bool is_empty(Entry *entry);

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
  free(table->entries);
  free(table);
}

void HT_set(HashTable *table, Key key, Value value) {
  if (table->count + 1 > table->capacity * TABLE_MAX_LOAD) {
    int capacity = GROW_CAPACITY(table->capacity);
    adjust_capacity(table, capacity);
  }

  Entry *entry = find_entry(table->entries, table->capacity, &key);
  if (is_empty(entry) && entry->value != TOMBSTONE_VALUE)
    table->count++;

  entry->key = key;
  entry->value = value;
}

Value HT_get(HashTable *table, Key key) {
  Entry *entry = find_entry(table->entries, table->capacity, &key);
  return entry == NULL ? NULL : entry->value;
}

void HT_delete(HashTable *table, Key key) {
  Entry *entry = find_entry(table->entries, table->capacity, &key);
  if (!is_empty(entry)) {
    entry->key.data = NULL;
    entry->key.size = 0;
    entry->value = TOMBSTONE_VALUE;
  }
}

static Entry *find_entry(Entry *entries, int capacity, Key *key) {
  int index = hash_key(*key) % capacity;
  Entry *tombstone = NULL;

  while (true) {
    Entry *entry = &entries[index];
    bool empty = is_empty(entry);
    if ((empty && entry->value != TOMBSTONE_VALUE) ||
        (!empty && key->size == entry->key.size && !memcmp(key->data, entry->key.data, key->size))) {
      return entry;
    } else if (empty && tombstone == NULL) {
      tombstone = entry;
    }

    index = (index + 1) % capacity;
  }

  return tombstone;
}

static void adjust_capacity(HashTable *table, int capacity) {
  Entry *entries = malloc(sizeof(Entry) * capacity);
  for (int i = 0; i < capacity; i++) {
    Entry *entry = &entries[i];
    entry->key = (Key){NULL, 0};
    entry->value = NULL;
  }

  table->count = 0;

  for (int i = 0; i < table->capacity; i++) {
    Entry *entry = &table->entries[i];
    if (is_empty(entry))
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
static uint32_t hash_key(Key key) {
  uint32_t hash = 2166136261u;
  for (size_t i = 0; i < key.size; i++) {
    hash ^= key.data[i];
    hash *= 16777619;
  }
  return hash;
}

static inline bool is_empty(Entry *entry) { return entry->key.data == NULL; }
