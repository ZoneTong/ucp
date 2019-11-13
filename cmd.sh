function clib(){
    go build -buildmode=c-archive -o dist/ucp.a
    gcc -o a.out tests/ucp_test/main.c dist/ucp.a
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

$@