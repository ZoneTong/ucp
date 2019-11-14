#include "../../dist/mudp.h"

#include <stdio.h>
#include <string.h>

int main() {
    char *config = "client.json";
    printf("%s", mudpInit(config));

    char *data= "data1zht";
    mudpResponse r = mudpSend("c1",  data, strlen(data));
    printf("%d %s\n", r.n, r.error);
    mudpReleaseResponse(r);

    data = "127.0.0.1";
    r = mudpSend("c1",  data, strlen(data));
    printf("%d %s\n", r.n,r.error);
    mudpReleaseResponse(r);

    int len = 50;
    char *rdata = (char *)malloc(len);
    r = mudpRecv("c1", rdata, len );
    printf("recvd %d,%s,%s\n",r.n,r.error, rdata);
    mudpReleaseResponse(r);
    
    printf("%s", mudpClose());
    return 0;
}