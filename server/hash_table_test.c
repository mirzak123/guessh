#include "hash_table.h"
#include <assert.h>
#include <string.h>

void test_hash_table(void) {
  Value value;
  char *value_str = "Hello world";
  int value_int = 1000, count = 0;

  HashTable *table = HT_create();
  assert(table != NULL);

  HT_set(table, KEY("key"), &value_int);
  count++;
  value = HT_get((table), KEY("key"));
  assert(*(int *)value == value_int);
  assert(table->capacity == 8);

  for (int i = count + 1; i < 10; i++) {
    HT_set(table, KEY(i), value_str);
    count++;
    value = HT_get(table, KEY(i));
    assert(!strcmp((char *)value, value_str));

    if (i <= 6)
      assert(table->capacity == 8);
    else
      assert(table->capacity == 16);
  }

  int x = 7;
  HT_set(table, KEY(x), value_str);
  assert(value_str == HT_get(table, KEY(x)));
  HT_delete(table, KEY(x));
  assert(NULL == HT_get(table, KEY(x)));

  // Find nonexistent
  value = HT_get(table, KEY("nonexistent"));
  assert(value == NULL);
  assert(table->capacity == 16);

  HT_destroy(table, NULL);
}

void test_call_HT_delete_on_empty_hash_table(void) {
  HashTable *ht = HT_create();
  HT_delete(ht, KEY(1));
}
