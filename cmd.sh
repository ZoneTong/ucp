go build -buildmode=c-archive -o dist/ucp.a
gcc -o a.out tests/ucp_test/main.c dist/ucp.a
./a.out