---
layout: post
title: "Brand new replay mode"
date: 2016-4-4 00:00:00 -0300
---
# Replay Mode

## Case of use
Most of the attacks or interesting events happening in a DNS server are **impossible to anticipate.** So, sooner than later you will miss an interesting event in your servers.

## Solution
Usually to solve this type of problems you will use *tcpreplay*, but in this case as the module might be in the same network, using tcpreplay will impact the network if the packets are injected with the same addresses. Or, if the addresses are appropriately changed, it  will cause a loss of information. That is why we just added to fievel the capability to retransmit an event stored in a *pcap* file.


## How to use it.
We tried to keep the configuration file as simple as possible, there are three fields you have to modify in order to have this new feature working:

```javascript
{
  ...
  "Capture" : {
    "InputMethod" : "capture_file",
    "InputSource" : "/path/to/the/pcap_file",
    "Speed" : n,
    ...
  },
  ...
}
```
*Sample of a configuration file.*

* **InputMethod:** You tell the system you want to read the data from a capture file.
* **InputSource:** The path to the pcap file you want to read from.
* **Speed:** This option allows you to set the speed the system will read the input. There are three different behaviors for this value:
 * **n = 0**: The system will read the data as fast as possible, no delay.
 * **n = 1**: The system will read the data with the same delay specified by the timestamps in the capture file.
 * **n > 0 && n != 1**: The system will read the data *n* times faster that the time specified in the capture file. This will speed up the reading if *n > 1* or speed it down if *n < 1.* For instance, if you set n = 2 the data will be read 2 times faster, but it you set *n = 0.25* it will take 4 times the original time to retransmit the data.

## Implementation details.
The delay was introduced by replacing the use of *pcap_loop* by our own loop. This loop includes calls to *nanosleep* in order to provide the pauses in the execution. As we are using *nanosleep* do not expect a better resolution than nanoseconds in the speed of the replay.
