# RaTA DNS - Real Time Analysis over DNS Packet Streams

RaTA DNS is a real time monitoring framework for multiple DNS servers, which is currently in development phase. The system's architecture is modular, to allow for easy extension and optimization of each component.

The system consists of a chain of packet processors, which starts in the network stream capture, goes through a serializer, then by a set of preliminary reducers, each of which sends its result through a different channel to an aggregator. The latter is responsible for generating the information needed to finally be consumed by the web visualization.

RaTA DNS modules are:

[RaTA DNS - Fievel](https://github.com/niclabs/ratadns-fievel) (serializer and preliminary reducer)

[RaTA DNS - Gopher](https://github.com/niclabs/ratadns-gopher) (aggregator)

[RaTA DNS - Remy](https://github.com/niclabs/ratadns-remy) (web visualization)

## Fievel - Monitor and Preliminary Reducer

Fievel is the first step in the pipeline RaTA DNS. Captures the data (DNS packets) coming from the network interface, a .pcap file or standard input. After the capture, the data goes through various preprocessing stages called Preliminary Reducers (PreR), that has as goal reduce the quantity of information to send to the next module.

Fievel is written in C, having each PreR as a Lua script, which are easily extensible.

To work, a window size and a preliminar reducers list is configured. Also, each PreR has its own configuration section, where it is defined which channel will be used to send the already preprocessed data. These channels may be the standard output, a file or a Redis channel.

Multiple PreR are already coded: an answers/queries per second calculator, a domain names counter (with and without clients IPs) and a queries summary grouped by host aggregator.

## Gopher - Aggregator Module

Gopher listens several Redis channels, one for each PreR. Multiple monitors (Fievel instances) send their data through the same channels. Gopher extracts the data, and start to process it. With this, aggregated data may be queried by the visualization.

Gopher works with a time window aggregation span defined in minutes. For example if the time window is of 1 minute, only stores and reports aggregated data of the last minute. Currently Gopher has implemented 3 Event Processors (different ways of aggregate the data incoming from Fievel) which are: queries/answers per second, information about queries made for each DNS record type existing and domains most queried with information about the clients that query it.

Gopher is written in Python and can easily be configured to add/remove event processors.

## Remy - Web Visualization

The final step is to visualize the data produced by Gopher. Currently Remy is in development phase as a HTML5 web application that shows the data delivered by one instance of Gopher (that can be receiving data from multiple DNS servers).

Remy shows three different visualizations corresponding to each event processor currently implemented in Gopher mentioned before.
