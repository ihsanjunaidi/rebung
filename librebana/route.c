/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

#include <arpa/inet.h>
#include <sys/queue.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/sysctl.h>
#include <sys/ioctl.h>
#include <net/route.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <net/if_types.h>
#include <net/if_var.h>
#include <netinet/in.h>
#include <netinet/in_var.h>
#include <netinet6/in6_var.h>
#include <netinet6/nd6.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "rebana.h"

const size_t    buflen = sizeof(struct rt_msghdr) + 512;

size_t  alignSa(size_t);

int
addRoute(struct sockaddr *dst, struct sockaddr *mask, struct sockaddr *gw)
{
    int                 s;
    char                *buf;
    struct sockaddr_in6 *sin6;
    struct rt_msghdr    *rtm;

    if ((s = socket(AF_ROUTE, SOCK_RAW, 0)) == -1)
        return(-1);

    /* fill in request header */
    buf = calloc(0, sizeof(char) * buflen);
    rtm = (struct rt_msghdr *) buf;
    rtm->rtm_version = RTM_VERSION;
    rtm->rtm_flags = RTF_UP | RTF_GATEWAY;
    rtm->rtm_addrs = RTA_DST | RTA_GATEWAY | RTA_NETMASK;
    rtm->rtm_type  = RTM_ADD;
    rtm->rtm_pid   = getpid();
    rtm->rtm_seq   = 0;

    /* take the dst AFI and assume the rest are of the same AFI */
    if (dst->sa_family == AF_INET6) {
        sin6 = (struct sockaddr_in6 *) (rtm + 1);
        memcpy(sin6, dst, sizeof(struct sockaddr_in6));

        sin6 = (struct sockaddr_in6 *) ((char *) sin6 +
                alignSa(sizeof(struct sockaddr_in6)));
        memcpy(sin6, gw, sizeof(struct sockaddr_in6));

        sin6 = (struct sockaddr_in6 *) ((char *) sin6 +
                alignSa(sizeof(struct sockaddr_in6)));
        memcpy(sin6, mask, sizeof(struct sockaddr_in6));

        sin6 = (struct sockaddr_in6 *) ((char *) sin6 +
                alignSa(sizeof(struct sockaddr_in6)));

        rtm->rtm_msglen = (char *) sin6 - buf;
    }

    if ((write(s, rtm, rtm->rtm_msglen)) == -1) {

        /* ignore route if exist */
        if (errno != EEXIST) {
            close(s);
            return(-1);
        }
    }

    close(s);

    return(0);
}

int
delRoute(struct sockaddr *dst, struct sockaddr *mask, struct sockaddr *gw)
{
    int                 s;
    char                *buf;
    struct sockaddr_in6 *sin6;
    struct rt_msghdr    *rtm;

    if ((s = socket(AF_ROUTE, SOCK_RAW, 0)) == -1) {
        return(-1);
    }

    /* fill in request header */
    buf = calloc(0, sizeof(char) * buflen);
    rtm = (struct rt_msghdr *) buf;
    rtm->rtm_version = RTM_VERSION;
    rtm->rtm_flags = RTF_UP | RTF_GATEWAY;
    rtm->rtm_addrs = RTA_DST | RTA_GATEWAY | RTA_NETMASK;
    rtm->rtm_type  = RTM_DELETE;
    rtm->rtm_pid   = getpid();
    rtm->rtm_seq   = 0;

    /* take the dst AFI and assume the rest are of the same AFI */
    if (dst->sa_family == AF_INET6) {
        sin6 = (struct sockaddr_in6 *) (rtm + 1);
        memcpy(sin6, dst, sizeof(struct sockaddr_in6));

        sin6 = (struct sockaddr_in6 *) ((char *) sin6 +
                alignSa(sizeof(struct sockaddr_in6)));
        memcpy(sin6, gw, sizeof(struct sockaddr_in6));

        sin6 = (struct sockaddr_in6 *) ((char *) sin6 +
                alignSa(sizeof(struct sockaddr_in6)));
        memcpy(sin6, mask, sizeof(struct sockaddr_in6));

        sin6 = (struct sockaddr_in6 *) ((char *) sin6 +
                alignSa(sizeof(struct sockaddr_in6)));

        rtm->rtm_msglen = (char *) sin6 - buf;
    }

    if ((write(s, rtm, rtm->rtm_msglen)) == -1) {

        /* ignore if it does not exist */
        if (errno != ESRCH) {
            close(s);
            return(-1);
        }
    }

    close(s);

    return(0);
}

size_t
alignSa(size_t s)
{
    return (1 + (((s) - 1) | (sizeof(size_t) - 1)));
}
