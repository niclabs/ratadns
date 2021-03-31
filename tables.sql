CREATE TABLE IF NOT EXISTS DNS_LOG (
  DnsDate Date,
  timestamp DateTime,
  Server String,
  IPVersion UInt8,
  IP String,
  Protocol FixedString(3),
  QR UInt8,
  OpCode UInt8,
  Class UInt16,
  Type UInt16,
  Edns0Present UInt8,
  DoBit UInt8,
  ResponseCode UInt8,
  Question String,
  Size UInt16,
  IPdst String
) engine=MergeTree(DnsDate, (timestamp, Server), 8192);

-- View for top queried domains
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_DOMAIN_COUNT
ENGINE=SummingMergeTree(DnsDate, (t, Server, Question), 8192, c) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, Question, count(*) as c FROM DNS_LOG WHERE QR=0 GROUP BY DnsDate, t, Server, Question;

-- View for unique domain count
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_DOMAIN_UNIQUE
ENGINE=AggregatingMergeTree(DnsDate, (t, Server), 8192) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, uniqState(Question) AS UniqueDnsCount FROM DNS_LOG WHERE QR=0 GROUP BY Server, DnsDate, t;

-- View for count by protocol
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_PROTOCOL
ENGINE=SummingMergeTree(DnsDate, (t, Server, Protocol), 8192, (c)) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, Protocol, count(*) as c FROM DNS_LOG GROUP BY Server, DnsDate, t, Protocol;

-- View with packet sizes
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_GENERAL_AGGREGATIONS
ENGINE=AggregatingMergeTree(DnsDate, (t, Server), 8192) AS
SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, sumState(Size) AS TotalSize, avgState(Size) AS AverageSize FROM DNS_LOG GROUP BY Server, DnsDate, t;

-- View with edns information
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_EDNS
ENGINE=AggregatingMergeTree(DnsDate, (t, Server), 8192) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, sumState(Edns0Present) as EdnsCount, sumState(DoBit) as DoBitCount FROM DNS_LOG WHERE QR=0 GROUP BY Server, DnsDate, t;

-- View wih query OpCode
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_OPCODE
ENGINE=SummingMergeTree(DnsDate, (t, Server, OpCode), 8192, c) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, OpCode, count(*) as c FROM DNS_LOG WHERE QR=0 GROUP BY Server, DnsDate, t, OpCode;

-- View with Query Types
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_TYPE
ENGINE=SummingMergeTree(DnsDate, (t, Server, Type), 8192, c) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, Type, count(*) as c FROM DNS_LOG WHERE QR=0 GROUP BY Server, DnsDate, t, Type;

-- View with Query Class
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_CLASS
ENGINE=SummingMergeTree(DnsDate, (t, Server, Class), 8192, c) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, Class, count(*) as c FROM DNS_LOG WHERE QR=0 GROUP BY Server, DnsDate, t, Class;

-- View with query responses
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_RESPONSECODE
ENGINE=SummingMergeTree(DnsDate, (t, Server, ResponseCode), 8192, c) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, ResponseCode, count(*) as c FROM DNS_LOG WHERE QR=1 GROUP BY Server, DnsDate, t, ResponseCode;

-- View with IP Prefix
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_IP_MASK
ENGINE=SummingMergeTree(DnsDate, (t, Server, IPVersion, IP), 8192, c) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, IPVersion, IP, count(*) as c FROM DNS_LOG WHERE QR=0 GROUP BY Server, DnsDate, t, IPVersion, IP;

-- View with IP servers
CREATE MATERIALIZED VIEW IF NOT EXISTS DNS_IP_SERVER
ENGINE=SummingMergeTree(DnsDate, (t, Server, IPVersion, IPdst), 8192, c) AS
  SELECT DnsDate, toStartOfMinute(timestamp) as t, Server, IPVersion, IPdst, count(*) as c FROM DNS_LOG WHERE QR=0 GROUP BY Server, DnsDate, t, IPVersion, IPdst;

