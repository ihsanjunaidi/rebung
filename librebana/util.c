/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <netdb.h>
#include <stdlib.h>
#include <string.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <net/if_types.h>

#include "rebana.h"

#include <stdio.h>

int
getAddrstr(char *addrs, struct sockaddr *sa)
{
    char                *s;
    void                *a;

    s = addrs;
    switch (sa->sa_family) {
        case AF_INET:
                a = &((struct sockaddr_in *) sa)->sin_addr;
                if (!a)
                    strlcpy(s, "no address", INET_ADDRSTRLEN);
                else
                    inet_ntop(sa->sa_family, a, s, INET_ADDRSTRLEN);

                break;

        case AF_INET6:
                a = &((struct sockaddr_in6 *) sa)->sin6_addr;
                if (!a)
                    strlcpy(s, "no address", INET6_ADDRSTRLEN);
                else
                    inet_ntop(sa->sa_family, a, s, INET6_ADDRSTRLEN);

                break;

        case AF_LOCAL:
        default:
                return(-1);
    }

    return(0);
}

int
getAddr(const char *ifa, struct sockaddr *sa)
{
    int             err, cnt;
    struct addrinfo hints, *res, *r;

    err = cnt = 0;

    memset(&hints, 0, sizeof(hints));
    hints.ai_socktype = SOCK_DGRAM;
    hints.ai_family = AF_UNSPEC;
    hints.ai_flags  = AI_NUMERICHOST;
    if ((err = getaddrinfo(ifa, NULL, &hints, &res)) == -1)
        return(-1);

    /* getaddrinfo() should return only 1 result per AF/socktype */
    for (r = res; r; r = r->ai_next) {
        if (r->ai_family == AF_INET)
            memcpy(sa, r->ai_addr, sizeof(struct sockaddr_in));
        else if (r->ai_family == AF_INET6)
            memcpy(sa, r->ai_addr, sizeof(struct sockaddr_in6));
    }

    freeaddrinfo(res);

    return(0);
}

int
getMask(uint8_t plen, struct sockaddr *sa)
{
    uint8_t             i, *s;

    switch (sa->sa_family) {
        case AF_INET:
            s = (uint8_t *) &((struct sockaddr_in *)sa)->sin_addr;
            for (i = 0; i < 4; i++, s++, plen -= 8) {
                if (plen >= 8) {
                    *s = 0xff;
                    continue;
                }

                *s = 0x00;
                break;
            }

            break;

        case AF_INET6:
            s = (uint8_t *) &((struct sockaddr_in6 *)sa)->sin6_addr;
            for (i = 0; i < 16; i++, s++, plen -= 8) {
                if (plen >= 8) {
                    *s = 0xff;
                    continue;
                }

                *s = 0x00;
                break;
            }

            break;
    }

    return(0);
}

int
getCidr(struct sockaddr *sa)
{
    uint8_t             i, plen, *s;

    plen = 0;

    /*all zeros netmask */
    if (sa->sa_len == 0)
        return(plen);

    switch (sa->sa_family) {
        case AF_INET:

            break;

        case AF_INET6:
            s = (uint8_t *) &((struct sockaddr_in6 *)sa)->sin6_addr;
            if (*s == 0)
                break;

            for (i = 0; ((i < 16) && (*s == 0xff)); i++, s++)
                plen += 8;

            break;

        default:
                return(-1);
    }

    return(plen);
}

int
getPort(struct sockaddr *sa)
{
    switch (sa->sa_family) {
        case AF_INET:
                return(ntohs(((struct sockaddr_in *) sa)->sin_port));

        case AF_INET6:
                return(ntohs(((struct sockaddr_in6 *) sa)->sin6_port));
    }

    return(0);
}
