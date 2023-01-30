#ifndef GRUEL_JIT_H
#define GRUEL_JIT_H

#include <jit/jit.h>

jit_int is_jit_supported();
jit_long compile_opcodes(jit_long length, jit_long *code);
void free_function(jit_long func);
jit_long call_jit_function(jit_long function, jit_long args, jit_long buf, jit_long len);

#endif /* !GRUEL_JIT_H */