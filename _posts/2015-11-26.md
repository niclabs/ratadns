---
layout: post
title: "Speedy Meets Fievel"
date: 2015-11-26 00:00:00 -0300
---
# Real Time Analysis over DNS Packet Streams

RaTA DNS is a real time monitoring framework for multiple DNS servers, which is currently in development phase. The system's architecture is modular, to allow for easy extension and optimization of each component.

The system consists of a chain of packet processors, which starts in the network stream capture, goes through a serializer, then by a set of preliminary reducers, each of which sends its result through a different channel to an aggregator. The latter is responsible for generating the information needed to finally be consumed by the web visualization.

# Monitor

A RaTA DNS monitor consists of a series of tools that define a pipeline. This pipeline is used to preprocess the data captured from a single server. At the end of this preprocessing, the results are sent to the aggregator. The aggregator receives preprocessed data from multiple monitors. Generally, we describe a RaTA DNS monitor as a machine whose sole function is to provide this service. However, RaTA DNS monitor can run on the same machine that runs the DNS server. The decision on which option to choose depends on network's needs.

![Monitor DNS]({{site.baseurl}}/images/2015-11-26-monitor.png)

The tools that comprise the monitor are TCPDUMP, Speedy and Fievel. These will be described in the sections to come.

# Network Stream Capture

To capture the network stream, 3 cases are considered: Capture directly on the same computer that works as server, capture via switch level port mirroring or replay a previous packet capture. The first two cases correspond to real-time capture of the packet, and the third case is a repetition of events in the past, this with forensic purposes.


![Packet Capture]({{site.baseurl}}/images/2015-11-26-port-mirroring.png "Packet capture using switch level port mirroring.")


Capturing the DNS packet stream, either directly on the server machine or through port mirroring, it's performed with the help of tcpdump. It has a stable and widespread format, commonly called PCAP. In addition to dump the captured packets into the disk, TCPDUMP may generate a stream into stdout. It's this mechanism that is used in Speedy -our packet serializer- to do its job.


# Speedy, Packet Serializer

TCPDUMP output is redirected through a pipe to Speedy, our packet serializer. This module takes a network packet in the PCAP format and interprets it as a DNS packet, including IP and UDP headers. Then analyzes this packet and writes it as a JSON message with a well-defined format. Which parts of the packet are serialized can be chosen when launching the tool. What parts may be chosen?. You may choose the source and destination IP addresses, the DNS's queries section, the DNS's answers section and the authoritative name servers DNS's section.

The module was developed in the C programming language, with the help of several libraries: libpcap was used to read the packets, ldns to read them as DNS packets and json-c packages to serialize the data in the JSON format. Furthermore, it was developed with the use of multiple cores in mind, by the use worker threads. The amount of worker threads can be passed as a parameter when running the program.

The serialization format is a length-prefixed JSON message, so message blocks can be quickly read. An example of Speedy output, both for a query to and an answer from the server is showed below. IP addresses are anonymized with a hash function due to privacy concerns:

### Query
```
128{"source":"0E402AE7","dest":"6DFA5FD7","id":"2e13","flags":"10","queries":[{"qname":"www.ejemplo.cl.","qtype":"1","qclass":"1"}]}
```

### Answer
```
584{"source":"6DFA5FD7","dest":"0E402AE7","id":"2e13","flags":"8010","queries":[{"qname":"www.ejemplo.cl.","qtype":"1","qclass":"1"}],"authorities":[{"name":"ejemplo.cl.","type":"2","class":"1","ns":["ejemplo.ns.cl."]},{"name":"ejemplo.cl.","type":"2","class":"1","ns":["ejemplo.ns.cl."]},{"name":"ejemplo.cl.","type":"2","class":"1","ns":["ejemplo.ns.cl."]},{"name":"ejemplo.cl.","type":"2","class":"1","ns":["ejemplo.ns.cl."]},{"name":"ejemplo.cl.","type":"2","class":"1","ns":["ejemplo.ns.cl."]},{"name":"ejemplo.cl.","type":"2","class":"1","ns":["ejemplo.ns.cl."]}]}
```

Speedy writes its data to the standard output, ready to be redirected to our next module: Fievel, the data preprocessing module.

# Fievel, Data Preprocessor

Fievel is RaTA DNS's data preprocessing module. Its objectives are to reduce the bandwidth occupied by the system and distribute the processing work across multiple monitors. It is a programmable module, where you can write different preliminary reducers, which are the real responsible for data processing.

![Packet Capture]({{site.baseurl}}/images/2015-11-26-multiple-monitors.png "Multiple RaTA DNS monitors may sent their data to the aggregator.")

To work, a window size and a preliminar reducers list is configured. Also, each preliminar reducer has its own configuration section, where it is defined which channel will be used to send the already preprocessed data. These channels may be the standard output, a file or a Redis channel.

In Fievel's mainloop, each JSON message sent by Speedy is received, and then it is transformed into a data structure known by preliminary reducers. They take the data structure and do their work. When all the packets that make a window are received, the mainloop tells the preliminar reducers to restart their state.

Multiple preliminar reducers are already coded. For example, a queries per second calculator, a malformed queries aggregator, a domain names counter, a queries summary grouped by host aggregator, etc. Some of them are showed below.

### Queries Name Counter:
```json
{
  "type": "QueriesNameCounter",
  "serverId": "server1",
  "data": {
    "www.ejemplo1.cl.": 5,
    "www.ejemplo2.cl.": 3,    
    "ns.ejemplo3.cl.": 4,
    "ejemplo4.cl.": 1,
  }
}
```

### Queries Per Second:
```json
{
  "type": "QueriesPerSecond",
  "serverId": "server1",
  "data": {
    "qps": 426.08993504359
  }
}
```


### Queries With Underscored Name:
```json
{
  "type": "QueriesWithUnderscoredName",
  "serverId": "server1",
  "data": {
    "malformed_query.domainname.cl.": [
      {
        "query": {
          "qclass": "1",
          "qtype": "1",
          "qname": "MALFORMED_QUERY.domainname.cl."
        },
        "sender": "be347435",
        "server": "c8107010"
      }
    ]
  }
}
```

# Jerry, Aggregator and Information Visualization

Jerry keeps listening several Redis channels, one for each PreR. Multiple monitors send their data through the same channels. Jerry extracts the data, and start to process it. With this, aggregated data may be queried by the visualization. More details will came in the future, with its own blog entry.

# RaTA DNS as a Forensic Tool

One of the RaTA DNS objectives is to study the behaviour that attackers had with the monitored services, somewhere in the past. With this, attacks patterns may be studied to know how to react in the future. 

To do this, it is required to permanently store DNS packets. Once this files are obtained, the tcpreplay tool is used to repeat the attack. With this tool, you can replay the captured network traffic, configuring even how many times faster do you want to repeat the simulation.

More details on this will be discussed later, too.
