#include "../../dist/common.h"

#include <stdio.h>
#include <string.h>

int main() {
    char *config = "tmp.json";
    initSDK(config);

    char *data= "data1zht";
    sentResponse r = sendData("a1", "head", data, strlen(data));
    printf("%d %s\n", r.r0, r.r1);
    releaseSentResponse(r);

    data = "127.0.0.1";
    r = sendData("a1", "head", data, strlen(data));
    printf("%d %s\n", r.r0,r.r1);
    releaseSentResponse(r);

    recvdResponse r1 = recvData("a1", "head");
    printf("recvd %d,%s,%s\n",r1.r0,r1.r1,r1.r2);
    releaseRecvdResponse(r1);
    closeSDK(NULL);
    return 0;
}