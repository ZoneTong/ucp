#include "../../dist/common.h"

#include <stdio.h>
#include <string.h>

int main() {
    char *config = "config.json";
    multipleInit(config);

    char *data= "data1zht";
    sentResponse r = multipleSend("c1",  data, strlen(data));
    printf("%d %s\n", r.r0, r.r1);
    releaseSentResponse(r);

    data = "127.0.0.1";
    r = multipleSend("c1",  data, strlen(data));
    printf("%d %s\n", r.r0,r.r1);
    releaseSentResponse(r);

    recvdResponse r1 = multipleRecv("c1" );
    printf("recvd %d,%s,%s\n",r1.r0,r1.r1,r1.r2);
    releaseRecvdResponse(r1);
    multipleClose();
    return 0;
}