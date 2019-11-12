go build -buildmode=c-archive -o number.a
gcc -o a.out cmain/number_test.c number.a
./a.out