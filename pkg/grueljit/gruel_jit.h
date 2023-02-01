#ifndef GRUEL_JIT_H
#define GRUEL_JIT_H

#include <jit/jit.h>

enum Type {
  GTYPE_PARENTHESIS = 0,
  GTYPE_BOOL,
  GTYPE_INT,
  GTYPE_FLOAT,
  GTYPE_STRING,
  GTYPE_SYMBOL,
};

typedef struct {
  jit_long ptr;
  jit_long len;
} go_string;

jit_int is_jit_supported();
jit_long compile_opcodes(jit_long length, jit_long *code, jit_long argc,
                         char *argv);
void free_function(jit_long func);
jit_long call_jit_function(jit_long function, jit_long args);

#endif /* !GRUEL_JIT_H */