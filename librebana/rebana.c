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
#include <ifaddrs.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "rebana.h"

#include <stdio.h>

int
createInterface(const char *ifname, const char *tunaddr, const char *tundest)
{
    int                 s, res = 0, err = 0;
    struct ifreq        ifr;
    struct in_aliasreq  inr;
    struct sockaddr_in  *sints, *sintc;

    if ((s = socket(AF_INET, SOCK_DGRAM, 0)) == -1)
        return(-1);

    memset(&ifr, 0, sizeof(ifr));
    strlcpy(ifr.ifr_name, ifname, IFNAMSIZ);

    if ((res = ioctl(s, SIOCIFCREATE, &ifr)) == -1) {
        err = 1;
        if ((errno == EEXIST) && ((res = ioctl(s, SIOCIFDESTROY, &ifr)) == 0)) {
            if ((res = ioctl(s, SIOCIFCREATE, &ifr)) == 0)
                err = 0;
        }
    }

    /* tun tunnel does not need assignment of point-to-point address */
    if (strnstr(ifname, "tun", IFNAMSIZ))
        return 0;

    sints = malloc(sizeof(struct sockaddr_in));
    sintc = malloc(sizeof(struct sockaddr_in));
    memset(sints, 0, sizeof(struct sockaddr_in));
    memset(sintc, 0, sizeof(struct sockaddr_in));
    getAddr(tunaddr, (struct sockaddr *) sints);
    getAddr(tundest, (struct sockaddr *) sintc);

    memset(&inr, 0, sizeof(inr));
    strlcpy(inr.ifra_name, ifname, IFNAMSIZ);
    memcpy(&inr.ifra_addr, sints, sizeof(struct sockaddr_in));
    memcpy(&inr.ifra_dstaddr, sintc, sizeof(struct sockaddr_in));

    if ((res = ioctl(s, SIOCSIFPHYADDR, &inr)) == -1) {
        err = 1;
    }

    close(s);
    free(sints);
    free(sintc);

    return 0;
}

int
createAddress(const char *ifname, const char *in6addr, uint8_t in6plen)
{
    int                 s;
    struct in6_aliasreq inr6;
    struct sockaddr_in6 *sin6s, *sin6m;

    sin6s = malloc(sizeof(struct sockaddr_in6));
    memset(sin6s, 0, sizeof(struct sockaddr_in6));
    getAddr(in6addr, (struct sockaddr *) sin6s);

    sin6m = malloc(sizeof(struct sockaddr_in6));
    memset(sin6m, 0, sizeof(struct sockaddr_in6));
    sin6m->sin6_len    = sizeof(struct sockaddr_in6);
    sin6m->sin6_family = AF_INET6;
    getMask(in6plen, (struct sockaddr *) sin6m);

    if ((s = socket(AF_INET6, SOCK_DGRAM, 0)) == -1)
        return -1;

    memset(&inr6, 0, sizeof(inr6));
    strlcpy(inr6.ifra_name, ifname, IFNAMSIZ);
    inr6.ifra_lifetime.ia6t_vltime = ND6_INFINITE_LIFETIME;
    inr6.ifra_lifetime.ia6t_pltime = ND6_INFINITE_LIFETIME;

    memcpy(&inr6.ifra_addr, sin6s, sizeof(struct sockaddr_in6));
    memcpy(&inr6.ifra_prefixmask, sin6m, sizeof(struct sockaddr_in6));

    if ((ioctl(s, SIOCAIFADDR_IN6, &inr6)) == -1)
        return -2;

    close(s);
    free(sin6s);
    free(sin6m);

    return 0;
}

int
deleteInterface(const char *ifname)
{
    int                 s;
    struct ifreq        ifr;

    if ((s = socket(AF_INET, SOCK_DGRAM, 0)) == -1)
        return(-1);

    memset(&ifr, 0, sizeof(ifr));
    strlcpy(ifr.ifr_name, ifname, IFNAMSIZ);
    if (ioctl(s, SIOCIFDESTROY, &ifr) == -1)
        return -1;

    close(s);
    return 0;
}

int
createRoute(const char *rt6dest, uint8_t rt6plen, const char *in6dest)
{
    struct sockaddr_in6 *sin6rt, *sin6rm, *sin6c;

    sin6rt = malloc(sizeof(struct sockaddr_in6));
    sin6c  = malloc(sizeof(struct sockaddr_in6));
    getAddr(rt6dest, (struct sockaddr *) sin6rt);
    getAddr(in6dest, (struct sockaddr *) sin6c);

    sin6rm = malloc(sizeof(struct sockaddr_in6));
    memset(sin6rm, 0, sizeof(struct sockaddr_in6));
    sin6rm->sin6_len    = sizeof(struct sockaddr_in6);
    sin6rm->sin6_family = AF_INET6;
    getMask(rt6plen, (struct sockaddr *) sin6rm);

    if (addRoute((struct sockaddr *) sin6rt, (struct sockaddr *) sin6rm,
                (struct sockaddr *) sin6c) == -1)
        return -1;

    free(sin6rt);
    free(sin6rm);
    free(sin6c);

    return 0;
}

int
deleteRoute(const char *rt6dest, uint8_t rt6plen, const char *in6dest)
{
    struct sockaddr_in6 *sin6rt, *sin6rm, *sin6c;

    sin6rt = malloc(sizeof(struct sockaddr_in6));
    sin6c  = malloc(sizeof(struct sockaddr_in6));
    getAddr(rt6dest, (struct sockaddr *) sin6rt);
    getAddr(in6dest, (struct sockaddr *) sin6c);

    sin6rm = malloc(sizeof(struct sockaddr_in6));
    memset(sin6rm, 0, sizeof(struct sockaddr_in6));
    sin6rm->sin6_len    = sizeof(struct sockaddr_in6);
    sin6rm->sin6_family = AF_INET6;
    getMask(rt6plen, (struct sockaddr *) sin6rm);

    if (delRoute((struct sockaddr *) sin6rt, (struct sockaddr *) sin6rm,
                (struct sockaddr *) sin6c) == -1)
        return -1;

    free(sin6rt);
    free(sin6rm);
    free(sin6c);

    return 0;
}
