/* Links libdeg.a and imports one bill: argv = template rules bill. */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "libdeg.h"

static char *slurp(const char *path, size_t *len) {
  FILE *f = fopen(path, "rb");
  if (!f) { perror(path); exit(2); }
  fseek(f, 0, SEEK_END);
  long n = ftell(f);
  fseek(f, 0, SEEK_SET);
  char *buf = malloc(n + 1);
  if (fread(buf, 1, n, f) != (size_t)n) { perror(path); exit(2); }
  buf[n] = 0;
  fclose(f);
  if (len) *len = n;
  return buf;
}

int main(int argc, char **argv) {
  if (argc != 4) { fprintf(stderr, "usage: smoke template rules bill\n"); return 2; }
  size_t bill_len = 0;
  char *tpl = slurp(argv[1], NULL), *rules = slurp(argv[2], NULL);
  char *bill = slurp(argv[3], &bill_len);
  printf("contract=%d\n", deg_contract_version());
  char *out = deg_import(tpl, rules, "bill.csv", (uint8_t *)bill, bill_len);
  printf("%s\n", out);
  int ok = strstr(out, "\"ok\":true") != NULL;
  deg_free(out);
  out = deg_import("template: [", "", "bill.csv", NULL, 0);
  ok = ok && strstr(out, "\"ok\":false") != NULL;
  deg_free(out);
  return ok ? 0 : 1;
}
