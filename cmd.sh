function clib(){
    # 直接使用go生成的静态库
    # go build -buildmode=c-archive -o dist/ucp.a
    # gcc -o a.out tests/ucp_test/main.c dist/ucp.a
    # ./a.out

    # 将go生成的静态库包装一层, 使函数参数可读
    go build -buildmode=c-archive -o dist/libucp.a
    cd dist
    gcc -c mudp.c  -o mudp.o

    # 方法1: 在可执行程序时link
    # ar -crs libmudp.a mudp.o  #归档成静态库
    # cd ..
    # gcc -o a.out tests/ucp_test/main.c -Ldist -lmudp -lucp

    # 方法2: 先拆分go静态库成.o文件
    ar -xs libucp.a                     # 拆解.a成.o
    ar -rcs libmudp.a *.o               # 打包成.a
    rm *.o libucp.a '__.SYMDEF SORTED'  #删除中间文件
    cd ..
    gcc -o a.out tests/ucp_test/main.c -Ldist -lmudp  # 编译成可执行程序

    ./a.out
}

function client(){
    cd tests/client && go build
    ./client
}

function server(){
    cd tests/server && go build 
    ./server
}

function treedir(){
    tree -I "global|multiple|common|client|server|simple"
}

$@