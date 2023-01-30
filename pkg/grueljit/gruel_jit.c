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
    int type = code[pc] & 0xff;
    jit_long value = code[pc + 1];
    if (type == GTYPE_PARENTHESIS) {
      switch (value) {
        BINARY_OP(1, jit_insn_add);
        BINARY_OP(2, jit_insn_sub);
        BINARY_OP(3, jit_insn_mul);
        BINARY_OP(4, jit_insn_div);
        BINARY_OP(5, jit_insn_rem);
      }
    } else if (type == GTYPE_SYMBOL) {
      jit_context_destroy(context);
      return 0;
    } else {
      jit_constant_t c;
      switch (type) {
      case GTYPE_BOOL:
        c.type = jit_type_sys_bool;
        break;
      case GTYPE_FLOAT:
        c.type = jit_type_float64;
        break;
      case GTYPE_INT:
        c.type = jit_type_long;
        break;
      }
      c.un.long_value = value;
      code[sp] = (jit_long)jit_value_create_constant(function, &c);
      sp++;
    }
  }

  if (sp < 1) {
    jit_function_abandon(function);
    jit_context_destroy(context);
    return 0;
  }

  jit_value_t ret = (jit_value_t)code[sp - 1];

  // Stores float64 in long.
  if (jit_value_get_type(ret) == jit_type_float64) {
    if (jit_value_is_constant(ret)) {
      jit_float64 c = jit_value_get_float64_constant(ret);
      jit_constant_t cValue;
      cValue.type = jit_type_long;
      cValue.un.float64_value = c;
      ret = jit_value_create_constant(function, &cValue);
    } else {
      jit_value_t address = jit_insn_address_of(function, ret);
      ret = jit_insn_load_relative(function, address, 0, jit_type_long);
    }
  }

  jit_insn_return(function, ret);

  if (!jit_function_compile(function)) {
    jit_function_abandon(function);
    jit_context_destroy(context);
    return 0;
  }
  jit_context_build_end(context);
  return (jit_long)function;
}

void free_function(jit_long func) {
  jit_function_t f = (jit_function_t)func;
  jit_context_t context = jit_function_get_context(f);
  jit_context_destroy(context);
}
