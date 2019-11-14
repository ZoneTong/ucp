#ifndef _MUCP_H
#define _MUCP_H
#include "libucp.h"

#include <stdio.h>
#include <stdlib.h>

// c编译静态库
// 1. 编译*.o
//      gcc *.c -c -I../include
// 2. 打包静态库.a
//      ar rcs libMyTest.a *.o
// 3. 使用静态库
        // 第一种方法：
        // gcc + 源文件 + -L 静态库路径 + -l静态库名 + -I头文件目录 + -o 可执行文件名
        // gcc main.c -L lib -l MyTest -I include -o app
        // ./app

        // 第二种方法：
        // gcc + 源文件 + -I头文件 + libxxx.a + -o 可执行文件名
        // gcc main.c -I include lib/libMyTest.a -o app



// 2nd compile 
typedef struct
{
    int n;
    char* error;
} mudpResponse;

char* mudpInit( char *config);
mudpResponse mudpSend( char *tag, const char *buf, int buflen);
mudpResponse mudpRecv( char *tag, char *buf, int buflen);
char* mudpClose();
void mudpReleaseResponse (mudpResponse r);

#endif