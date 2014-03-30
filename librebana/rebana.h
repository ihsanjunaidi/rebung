/*
 * Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
 */

#ifndef _REBANA_H_
#define _REBANA_H_

#include <sys/socket.h>
#include <sys/types.h>

struct session {
    char                    *ifname;
    char                    *tunaddr;
    char                    *tundest;
    char                    *in6addr;
    char                    *in6dest;
    uint8_t                 in6plen;
    char                    *rt6dest;
    uint8_t                 rt6plen;
};

/* prototypes */
/* rebana.c */
int     createInterface(const char *, const char *, const char *);
int     createAddress(const char *, const char *, uint8_t);
int     createRoute(const char *, uint8_t, const char *);
int     deleteInterface(const char *);
int     deleteRoute(const char *, uint8_t, const char *);

/* route.c */
int     addRoute(struct sockaddr *, struct sockaddr *, struct sockaddr *);
int     delRoute(struct sockaddr *, struct sockaddr *, struct sockaddr *);

/* util.c */
int     getAddrstr(char *, struct sockaddr *);
int     getAddr(const char *, struct sockaddr *);
int     getMask(uint8_t, struct sockaddr *);
int     getPort(struct sockaddr *);
int     getCidr(struct sockaddr *);

#endif  /* _REBANA_H_ */
