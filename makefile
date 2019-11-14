first_is_default: exec2

all: lib2 exec1

golib:
	go build -buildmode=c-archive -o dist/libucp.a

mudp:
	cd dist && gcc -c mudp.c  -o mudp.o

lib1: golib mudp

	# 拆解 并重新合并
	cd dist && ar -xs libucp.a && ar -rcs libmudp.a *.o

	#删除中间文件
	cd dist && rm *.o '__.SYMDEF SORTED'    
	

lib2: golib mudp
	cd dist && rm -f libmudp.a && ar -crs libmudp.a mudp.o

# 仅依赖一个单独的库文件
exec1:	lib1
	gcc -o a.out tests/ucp_test/main.c -Ldist -lmudp

# 需要依赖多个库文件
exec2: lib2
	gcc -o a.out tests/ucp_test/main.c -Ldist -lmudp -lucp

.PHONY: clean
clean:
	rm -f a.out
	cd dist && rm -f *.a *.o