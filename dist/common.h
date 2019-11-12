#include "ucp.h"

#include <stdio.h>
#include <stdlib.h>

typedef struct multipleSend_return sentResponse;

void releaseSentResponse(sentResponse r)
{
    free(r.r1);
}

typedef struct multipleRecv_return recvdResponse;

void releaseRecvdResponse(recvdResponse r)
{
    free(r.r1);
    free(r.r2);
}