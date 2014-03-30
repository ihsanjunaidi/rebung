#
# Copyright (c) 2013 Ihsan Junaidi Ibrahim <ihsan.junaidi@gmail.com>
#

LIB=	rebana

SRCS=	rebana.c \
        route.c \
        util.c

LDADD=  -L${LOCALBASE}/lib -lutil

CFLAGS+= -Wall -I${.CURDIR} -I/usr/local/include
CFLAGS+= -Wstrict-prototypes -Wmissing-prototypes
CFLAGS+= -Wmissing-declarations
CFLAGS+= -Wshadow -Wpointer-arith -Wcast-qual
CFLAGS+= -Wsign-compare -Werror

.include <bsd.lib.mk>
.include <bsd.port.mk>
