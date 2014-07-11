// Sets up a Keyspace with replication settings suitable for testing on an unmodified default Cassandra install.
CREATE KEYSPACE mcp WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

USE phylactery;

// In the future I'll use a map type for the datacenters and map it to a struct in Go, but for now this works.
CREATE TABLE file_ledger (
  fid     text PRIMARY KEY,
  created timestamp,
  origin  text,
  ma01    boolean,
  tx01    boolean
);

// Secondary indexes to make some faster queries possible.
CREATE INDEX ma01_state ON mcp.file_ledger (ma01);
CREATE INDEX tx01_state ON mcp.file_ledger (tx01);

// Sample data.
INSERT INTO file_ledger (fid, created, origin, ma01, tx01) VALUES('1234.fid', dateof(now()), 'ma01', true, false);
INSERT INTO file_ledger (fid, created, origin, ma01, tx01) VALUES('5678.fid', dateof(now()), 'ma01', true, true);
