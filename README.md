Moss Tier [![Build Status](https://travis-ci.org/fuglede/mosstier.svg?branch=master)](https://travis-ci.org/fuglede/mosstier)
=========

The goal of this WIP repository is to replace https://mosstier.com by an open source alternative.


Installation
============

To set up the server, first make sure that you have `mysql` and `go` installations on relevant machine. Then, rather than `go get`ting (which is of course also possible), the fastest way to get started is probably get the current source code from GitHub using

    git clone git@github.com:fuglede/mosstier.git
    cd mosstier

Then, import the empty database from the schema in this repository, for instance by running

    mysql -u database_username -p < schema.sql

Now, to set up the Moss Tier installation, simply move the example configuration and edit its contents to reflect your own setup,

    mv config.json.example config.json

That's pretty much it; to test your setup, run

    go run *.go

and navigate your webserver to http://localhost:9090 (or whatever combination of host and port you decide to use).
