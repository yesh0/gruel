#include "gruel_jit.h"

jit_int is_jit_supported() {
  jit_init();
  return !jit_uses_interpreter();
}

jit_long call_jit_function(jit_long function, jit_long args, jit_long len) {
  if (function == 0 || (args == 0 && len != 0)) {
    return 0;
  }
  jit_function_t f = (jit_function_t)function;
  jit_long *parameters = (jit_long *)args;
  void *entry = jit_function_to_closure(f);
  return ((jit_long(*)(jit_long *))entry)(parameters);
}

#define BINARY_OP(opcode, func)                                                \
  case (opcode):                                                               \
    if (sp < 2) {                                                              \
      jit_context_destroy(context);                                            \
      return 0;                                                                \
    }                                                                          \
    sp--;                                                                      \
    code[sp - 1] = (jit_long)func(function, (jit_value_t)code[sp],             \
                                  (jit_value_t)code[sp - 1]);                  \
    break

#define UNARY_OP(opcode, func)                                                 \
  case (opcode):                                                               \
    if (sp < 1) {                                                              \
      jit_context_destroy(context);                                            \
      return 0;                                                                \
    }                                                                          \
    code[sp - 1] = (jit_long)func(function, (jit_value_t)code[sp - 1]);        \
    break

jit_long compile_opcodes(jit_long length, jit_long *code, jit_long argc,
                         char *argv) {
  jit_context_t context = jit_context_create();
  if (!context) {
    return 0;
  }
  jit_context_build_start(context);

  jit_type_t signature;
  jit_type_t paramType = jit_type_void_ptr;
  signature =
      jit_type_create_signature(jit_abi_cdecl, jit_type_long, &paramType, 1, 1);
  jit_function_t function = jit_function_create(context, signature);
  if (!function) {
    jit_context_destroy(context);
    return 0;
  }

  jit_value_t paramBase = jit_value_get_param(function, 0);

  int sp = 0;
  for (int pc = 0; pc < length; pc += 2) {
    int type = code[pc] & 0xff;
    jit_long value = code[pc + 1];
    if (type == GTYPE_PARENTHESIS) {
      switch (value) {
        //@start maintained by operators.go
        // `+`(2)
        BINARY_OP(0x01, jit_insn_add);
        // `-`(2)
        BINARY_OP(0x02, jit_insn_sub);
        // `-`(1)
        UNARY_OP (0x03, jit_insn_neg);
        // `*`(2)
        BINARY_OP(0x04, jit_insn_mul);
        // `/`(2)
        BINARY_OP(0x05, jit_insn_div);
        // `%`(2)
        BINARY_OP(0x06, jit_insn_rem);
        // `&`(2)
        BINARY_OP(0x07, jit_insn_and);
        // `|`(2)
        BINARY_OP(0x08, jit_insn_or);
        // `^`(2)
        BINARY_OP(0x09, jit_insn_xor);
        // `^`(1)
        UNARY_OP (0x0a, jit_insn_not);
        // `<<`(2)
        BINARY_OP(0x0b, jit_insn_shl);
        // `>>`(2)
        BINARY_OP(0x0c, jit_insn_shr);
        // `>>>`(2)
        BINARY_OP(0x0d, jit_insn_ushr);
        // `**`(2)
        BINARY_OP(0x10, jit_insn_pow);
        // `acos`(1)
        UNARY_OP (0x11, jit_insn_acos);
        // `asin`(1)
        UNARY_OP (0x12, jit_insn_asin);
        // `atan`(1)
        UNARY_OP (0x13, jit_insn_atan);
        // `atan2`(2)
        BINARY_OP(0x14, jit_insn_atan2);
        // `cos`(1)
        UNARY_OP (0x15, jit_insn_cos);
        // `cosh`(1)
        UNARY_OP (0x16, jit_insn_cosh);
        // `exp`(1)
        UNARY_OP (0x17, jit_insn_exp);
        // `log`(1)
        UNARY_OP (0x18, jit_insn_log);
        // `log10`(1)
        UNARY_OP (0x19, jit_insn_log10);
        // `pow`(2)
        BINARY_OP(0x1a, jit_insn_pow);
        // `sin`(1)
        UNARY_OP (0x1b, jit_insn_sin);
        // `sinh`(1)
        UNARY_OP (0x1c, jit_insn_sinh);
        // `sqrt`(1)
        UNARY_OP (0x1d, jit_insn_sqrt);
        // `tan`(1)
        UNARY_OP (0x1e, jit_insn_tan);
        // `tanh`(1)
        UNARY_OP (0x1f, jit_insn_tanh);
        //@end maintained by operators.go
      default:
        jit_context_destroy(context);
        return 0;
      }
    } else if (type == GTYPE_SYMBOL) {
      if (argv[value] == GTYPE_FLOAT) {
        code[sp] = (jit_long)jit_insn_load_relative(
            function, paramBase, value * 8, jit_type_float64);
      } else {
        code[sp] = (jit_long)jit_insn_load_relative(function, paramBase,
                                                    value * 8, jit_type_long);
      }
      sp++;
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
    // Marks that we are returning a float64.
    ((char *)code)[0] = 0xff;
  } else {
    ((char *)code)[0] = 0;
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
  if (func == 0) {
    return;
  }
  jit_function_t f = (jit_function_t)func;
  jit_context_t context = jit_function_get_context(f);
  jit_context_destroy(context);
}
