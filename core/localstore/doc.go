/*
Package localstore provides disk storage layer for Swarm Chunk persistence.
It uses swarm/shed abstractions.

The main type is DB which manages the storage by providing methods to
access and add Chunks and to manage their status.

Modes are abstractions that do specific changes to Chunks. There are three
mode types:

 - ModeGet, for Chunk access
 - ModePut, for adding Chunks to the database
 - ModeSet, for changing Chunk statuses

Every mode type has a corresponding type (Getter, Putter and Setter)
that provides adequate method to perform the opperation and that type
should be injected into localstore consumers instead the whole DB.
This provides more clear insight which operations consumer is performing
on the database.

Getters, Putters and Setters accept different get, put and set modes
to perform different actions. For example, ModeGet has two different
variables ModeGetRequest and ModeGetSync and two different Getters
can be constructed with them that are used when the chunk is requested
or when the chunk is synced as this two events are differently changing
the database.

Subscription methods are implemented for a specific purpose of
continuous iterations over Chunks that should be provided to
Push and Pull syncing.

DB implements an internal garbage collector that removes only synced
Chunks from the database based on their most recent access time.

Internally, DB stores Chunk data and any required information, such as
store and access timestamps in different shed indexes that can be
iterated on by garbage collector or subscriptions.
*/
package localstore
