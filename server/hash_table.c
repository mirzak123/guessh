#include "hash_table.h"
#include <_stdio.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define INITIAL_CAPACITY 8
#define TABLE_MAX_LOAD 0.75
#define GROW_CAPACITY(capacity) ((capacity) < INITIAL_CAPACITY ? INITIAL_CAPACITY : (capacity) * 2)

static Entry *find_entry(Entry *entries, int capacity, Key *key);
static void adjust_capacity(HashTable *table, int capacity);
static uint32_t hash_key(Key k);

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
  entry->key = key;
  entry->value = value;
}

Value HT_get(HashTable *table, Key key) {
  Entry *entry = find_entry(table->entries, table->capacity, &key);
  return entry == NULL ? NULL : entry->value;
}

static Entry *find_entry(Entry *entries, int capacity, Key *key) {
  int index = hash_key(*key) % capacity;

  while (true) {
    Entry *entry = &entries[index];
    if (entry->key.data == NULL || !strcmp((char *)key->data, (char *)entry->key.data)) {
      return entry;
    }

    index = (index + 1) % capacity;
  }
}

static void adjust_capacity(HashTable *table, int capacity) {
  Entry *entries = malloc(sizeof(Entry) * capacity);
  for (int i = 0; i < capacity; i++) {
    Entry *entry = &entries[i];
    entry->key = (Key){NULL, 0};
    entry->value = NULL;
  }

  for (int i = 0; i < table->capacity; i++) {
    Entry *entry = &table->entries[i];
    if (entry->key.data == NULL)
      continue;

    Entry *new_entry = find_entry(entries, capacity, &entry->key);
    new_entry->key = entry->key;
    new_entry->value = entry->value;
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
