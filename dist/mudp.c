#include "mudp.h"


// typedef struct multipleSend_return sentResponse;

// void releaseSentResponse(sentResponse r)
// {
//     free(r.r1);
// }

// typedef struct multipleRecv_return recvdResponse;

// void releaseRecvdResponse(recvdResponse r)
// {
//     free(r.r1);
// }


char* mudpInit( char *config){
    return multipleInit(config);
}

mudpResponse mudpSend( char *tag, const char *buf, int buflen){
    struct multipleSend_return r = multipleSend(tag, (char *)buf, buflen);
    mudpResponse resp;
    resp.n = r.r0;
    resp.error = r.r1;
    return resp;
}

mudpResponse mudpRecv( char *tag, char *buf, int buflen){
    struct multipleRecv_return r = multipleRecv(tag, buf, buflen);
    mudpResponse resp;
    resp.n = r.r0;
    resp.error = r.r1;
    return resp;
}

char* mudpClose(){
    return multipleClose();
}

void mudpReleaseResponse (mudpResponse r){
    if (r.error != NULL){
        free(r.error);
        r.error = NULL;
    }
}