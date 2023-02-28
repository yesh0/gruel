libjit: libjit/libjit.a

libjit/jit/.libs/libjit.a: .gitmodules
	cd libjit; ./bootstrap && ./configure && make

libjit/libjit.a: libjit/jit/.libs/libjit.a
	cp libjit/jit/.libs/libjit.a libjit/libjit.a

.PHONY: libjit
