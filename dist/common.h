#include "ucp.h"

#include <stdio.h>
#include <stdlib.h>

typedef struct sendData_return sentResponse;

void releaseSentResponse(sentResponse r)
{
    free(r.r1);
}

typedef struct recvData_return recvdResponse;

void releaseRecvdResponse(recvdResponse r)
{
    free(r.r1);
    free(r.r2);
}