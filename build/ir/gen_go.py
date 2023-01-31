#!/bin/python3

header = "libjit/include/jit/jit-insn.h"


def index_of(word, lines):
    """Index of sub string"""
    for i, line in enumerate(lines):
        if word in line:
            return i
    return -1


excludes = []
mapping = {
    "eq": "==",
    "ne": "!=",
    "lt": "<",
    "le": "<=",
    "gt": ">",
    "ge": ">=",
    "to_not_bool": "!",
}
aliases = {
    "==": ["=", "=="],
    "pow": ["pow", "**"],
}
wrapped = {
    "jit_insn_eq": "gruel_insn_eq",
    "jit_insn_ne": "gruel_insn_ne",
}


def main():
    """main"""
    with open(header, encoding="UTF-8") as f:
        opcode = 0x40
        functions = []
        lines = f.readlines()
        start = index_of("jit_insn_eq", lines)
        end = index_of("jit_insn_sign", lines) + 1
        i = start
        while i <= end:
            if not lines[i].startswith("jit_value_t"):
                i += 1
                continue
            full_name = lines[i].split(" ")[1].strip()
            name = full_name.split("_", 2)[2]
            if name in excludes:
                i += 1
                continue
            if name in mapping:
                name = mapping[name]
            if name.startswith("to_"):
                name = "->" + name[3:]
            if name.startswith("is_"):
                name = name[3:] + "?"
            argc = 1
            if "value2" in lines[i] or "value2" in lines[i + 1]:
                argc = 2
            if full_name in wrapped:
                full_name = wrapped[full_name]
            for alias in (aliases[name] if name in aliases else [name]):
                functions.append(
                    f"\"{alias}\": " +
                    f"[]Operator{{{{{hex(opcode)}, {argc}, nil, \"{full_name}\"}}}},"
                )
                opcode += 1
            i += 1
        print("\n".join(functions))


main()
