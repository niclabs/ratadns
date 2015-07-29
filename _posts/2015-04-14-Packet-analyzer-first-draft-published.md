---
layout: post
title: "Packet Analyzer first draft published"
date: 2015-04-14 15:00:00 -0300
---

#Dependencies

- [ldns](http://www.nlnetlabs.nl/projects/ldns/): DNS C library, used to parse DNS packets
- [json-c](https://github.com/json-c/json-c): JSON C library, used to generate the output
- [pkg-config](http://www.freedesktop.org/wiki/Software/pkg-config/) _(Optional)_: Helper tool used to locate and json-c library folders to link in Makefile. If you don't have `pkg-config` installed, you have to modify the Makefile according to [this site](https://github.com/json-c/json-c#linking-to-libjson-c).


#Capabilities

14-04-2015

- Read from libpcap stream
- Filter DNS packets (by UDP protocol & port 53)
- Parse IPv4 Header information
- Parse IPv6 Header information (partial)
- Parse DNS Queries section
