phylactery
==========

In D&D, a [phylactery](http://en.wikipedia.org/wiki/Phylactery) is a magic container used to store a [Lich's](http://en.wikipedia.org/wiki/Lich_(Dungeons_%26_Dragons)) soul.

![phylactery](http://www.quickmeme.com/img/57/57ab7a1d5ded6007388965c3cc076c12ad0a492a98f578c613854d4d40d90da7.jpg "phylactery")

This is mainly a proof-of-concept Go program for interacting with Cassandra.   Partially as a toy to learn Go with and partially to flesh out the design for a global file ledger abstraction to sync up geographically disparate object stores and to make different object store pluggable.

This is mainly to get the idea and notes out of my head.

Right now I am dealing with a globally namespaced, single MySQL master [MogielFS](https://code.google.com/p/mogilefs/) object store hosung ~10PB of data across 3 datacenters.

Looking to the future I'd rather each data center have it's own distinct MogileFS installation that looks to itself and maintains it's own replica of the data.

Then, I would have something external that keeps a list of all of the files that _should_ be available (this project, the Phylactery)

When a file is written to any data center, the client/service notifies the Phylactery and that same process notes the files are missing and grabs them.

Finally, there would be a proccess interacting with each data center's MogileFS installation to make sure it has all of the files listed in the Phylactery and take corrective action if needed (get the missing file from another datacenter)

This has the advantages of:

* Smaller failure domains and we are not beholden to a single datastore in one data center.
* Allowing each datacenter to work standalone and allow repair to happen offline.
* Make adding new datacenters easier and incremental.
* As we have different asset/object types we can have a finer grain control of what goes in what data center if needed.
* We can swap out MogileFS, and use Swift or Ceph or what have you and just hook it up to the Phylactery.  I love not being beholden to any one technology.

I use Cassandra to make sure the data is in every datacenter and since I do not care how long writes take, I can write a CL.ALL to assure all Cassandra nodes get the update.

To tinker:

First, [Install Cassandra](http://www.datastax.com/documentation/cassandra/2.0/cassandra/install/installDeb_t.html):

Installing via `apt-get` will start it automatically, you need not do any config changes.

Use `cqlsh` to connect to the local Cassandra instance and copy-paste from the `setup.cql` file.

Assuming you have Go installed and your `GOPATH` set grab the `cocql` dependancy:

```
go get github.com/gocql/gocql

```

Then, just build it
```
go build phylactery.go 
```

And run
```
./phylactery 
```

Here is some example usage:

List a file that is not in tx01
```
curl http://localhost:8080/file/bad/tx01
```

Toggle that file as now in tx01
```
curl -X POST -d "{\"Fid\":\"1234.fid\"}" http://localhost:8080//file/add/tx01
```

Add a new file to the ledger.
```
curl -X POST -d "{\"Fid\":\"1234.fid\",\"Origin\":\"ma01\",\"Ma01\":true,\"Tx01\":false}" http://localhost:8080/file/new
```

Obviously the schema would need to be improved and I'd like to use a map in Cassandra instead of each datacenter... but whatever, will get to it.

Moreover, I can build plugin in MogileFS that will write back other metadata data to Cassandra if I want to do some real analytics with [Spark/Shark](http://www.slideshare.net/planetcassandra/c-summit-2013-realtime-analytics-using-cassandra-spark-and-shark-by-evan-chan)
