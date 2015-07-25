---
layout: post
title: "RaTA-DNS Project Architecture"
date: 2015-07-24 18:00:00 -0300
---

RaTA-DNS is composed by three modules: a packet analyzer, a processing logic framework, and a data visualization module.

<center>
<img src={{site.urlbase}}/images/Arquitectura_trans.png width=80%>
</center>

Packet analyzer is our equivalent to the DSC collector tool. It takes a DNS packet stream and converts them into JSON objects (a lightweight data-interchange format), which are compressed and sent via a secure connection to a hub. How to capture packets depends on the needs of the DNS admins, but in the first instance `tcpdump` or even the same DSC collector can be used for this purpose. The default pipeline defined in
this module is as follows:

- Capture packets with tcpdump.
- Scan them with the packet stream analyzer.
- Compress them with LZ4.
- Send to a hub machine.

Same pipeline has to be used in each of the servers wanted to be analyzed.

When the data is received by the hub machine, it is decompressed and sent into a Kafka server.

Then, the Apache Storm framework read the data from the hub and distribute the work to multiple nodes, in order to obtain different types of statistics and aggregations defined by TLD administrators. Finally, aggregations and statistics are sent to a distributed database, such as Redis, in order to be displayed.

Visualization is an on-going work that will work in a distributed manner, with an HTML5 frontend. It is planned to use the [R programming language](http://www.r-project.org/) with the [Shiny](http://shiny.rstudio.com/) web framework.
