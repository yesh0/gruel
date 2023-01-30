#include "gruel_jit.h"

jit_int is_jit_supported() {
  jit_init();
  return !jit_uses_interpreter();
}

jit_long call_jit_function(jit_long function, jit_long args, jit_long buf,
                           jit_long len) {
  if (function == 0 || ((args == 0 || buf == 0) && len != 0)) {
    return 0;
  }
  jit_function_t f = (jit_function_t)function;
  jit_long *parameters = (jit_long *)args;
  void **list = (void **)buf;
  for (int i = 0; i < len; ++i) {
    list[i] = (void *)&parameters[i];
  }
  jit_long ret;
  jit_function_apply(f, list, &ret);
  return ret;
}

#define BINARY_OP(opcode, func)                                                \
  case (opcode):                                                               \
    if (sp < 2) {                                                              \
      jit_context_destroy(context);                                            \
      return 0;                                                                \
    }                                                                          \
    sp--;                                                                      \
    code[sp - 1] = (jit_long)func(function, (jit_value_t)code[sp - 1],         \
                                  (jit_value_t)code[sp]);                      \
    break

jit_long compile_opcodes(jit_long length, jit_long *code) {
  jit_context_t context = jit_context_create();
  if (!context) {
    return 0;
  }
  jit_context_build_start(context);

  jit_type_t signature;
  signature =
      jit_type_create_signature(jit_abi_cdecl, jit_type_long, NULL, 0, 1);
  jit_function_t function = jit_function_create(context, signature);
  if (!function) {
    jit_context_destroy(context);
    return 0;
  }

  int sp = 0;
  for (int pc = 0; pc < length; pc += 2) {
    jit_long type = code[pc];
    jit_long value = code[pc + 1];
    switch (type) {
    case 0:
      switch (value) {
        BINARY_OP(1, jit_insn_add);
        BINARY_OP(2, jit_insn_sub);
        BINARY_OP(3, jit_insn_mul);
        BINARY_OP(4, jit_insn_div);
        BINARY_OP(5, jit_insn_rem);
      }
      break;
    default:;
      jit_constant_t c;
      c.type = jit_type_long;
      c.un.long_value = value;
      code[sp] = (jit_long)jit_value_create_constant(function, &c);
      sp++;
      break;
    }
  }
  if (sp < 1) {
    return 0;
  }
  jit_insn_return(function, (jit_value_t)code[sp - 1]);
  if (!jit_function_compile(function)) {
    jit_function_abandon(function);
    jit_context_destroy(context);
    return 0;
  }
  jit_context_build_end(context);
  return (jit_long)function;
}

void free_function(jit_long func) {
  jit_function_t f = (jit_function_t) func;
  jit_context_t context = jit_function_get_context(f);
  jit_context_destroy(context);
}
