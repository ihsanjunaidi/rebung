rebung
======

Experimental prototype for a conceptual IPv6 6in4 tunneling web service using Go. Coded on FreeBSD 9 with Go 1.2 and CGo.

Contains a few components:
- rdbtool (populate/reset Redis DB with test data)
- ghazal (user database web service)
- rebana (6in4 tunnel manager web service)
- rebanats (6in4 tunnel host controller, linked to librebana via CGo)
- librebana (low-level C calls to kernel intf/routing table - FreeBSD-specific)
- rctlweb (admin control panel)
- rweb (primary user web)

Feedback if you like it enough. Thank you.
