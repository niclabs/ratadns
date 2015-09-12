---
layout: post
title: "A DNS Analysis Tale"
date: 2015-12-9 00:00:00 -0300
---

Since our last blog entry the architectural design of RaTA DNS has changed.
We needed to reduce the network traffic that our Packet Parser -currently named *Speedy*, after "Speedy Gonzalez"- was generating. To do this, we designed and implemented *Fievel*, a new module which's main purpose is to perform preliminar packet filtering-transformation.

<center>
<img src={{site.baseurl}}/images/monitor.png width=80%>
</center>

So, lets summarize what we are doing here. Think that we have a DNS server and we want to have a realtime monitoring tool. A way that doesn't add overhead to the server is to do port mirroring to an interface that is connected to an analyzer machine, just by configuring the switch. In this machine we have our RaTA DNS pipeline working: we are sniffing DNS packets with tcpdump, then we serialize the packets into a slightly-modified-JSON formatted message with the help of *Speedy*, following with the preliminar packet filtering-transformation given by *Fievel*. Fievel has several preliminar filtering-transformers (PFT) inside, each of these knows how to serialize a processed packet window. For example, a PFT may be configured to send the processed window directly to a visualization module or a data-aggregation module on another machine.

### What does *Fievel* exactly do? ###

To start *Fievel* we have to configure *L*, a list of PFTs, and *W*, its windows size parameter: the number of packets that the module is going to analyze before it reset the complete state of the program. Once a serialized DNS packet is received from *Speedy*, *Fievel*'s main loop send this packet to every preliminar filtering-transformer in *L*. When *Fievel* receive the *W*-th packet, it tell each PFT to serialize the processed window and to reset it state.

### And what is this PFT thing? ###

A PFT is an object that stores some state, knows how to process one packet at a time, knows what to do when Fievel received *W* packets, and knows how to reset its own state. 
A simple example of a PFT is our *Packets Per Seconds* showed below:

{% highlight python %}

from prer import PreR
import time

class PacketsPerSecond(PreR):
    def __init__(self, f):
        PreR.__init__(self, f)
        self.counter = 0
        self.start = time.time()

    def __call__(self, d):
        self.counter += 1

    def get_data(self):
        data = { 'pps' : self.counter / (time.time() - self.start) }
        return data

    def reset(self):
        self.counter = 0
        self.start = time.time()

{% endhighlight %}

So that's it! All our code is [publicly available](https://github.com/niclabs/ratadns-filters) under a MIT license. If you have any suggested PFT, on any suggestion in general don't hesitate to contact us!
