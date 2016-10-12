ipfs-see-all
============

ipfs-see-all is a diagnostics and recovery utility for ipfs repos. It has two
modes of operation and must be run with your local ipfs daemon shut down.

### lost-pins
The 'lost-pins' mode will search through all the ipfs objects in your
repo and look for objects that appear to be pinsets. It enumerates these
potential pinsets and compares the results with the pinset reported by your
nodes 'actual' pinset. If any pins are found in the loose objects that are not
found in the 'actual' pinset, they will be printed out. The printed out hashes
represent objects that were at one time pinned, but are no longer. This can
happen for one of two reasons, first (and hopefully most likely), these could
be objects that you have manually unpinned (via `ipfs pin rm`). Second, they
could be objects whose pins were lost by the ipfs repo [pinset
bug](https://github.com/ipfs/go-ipfs/pull/3273). If you suspect the pins
printed out were lost, first make sure you've updated to a [go-ipfs v0.4.4 or later](https://dist.ipfs.io/#go-ipfs),
then you can run `ipfs pin add` on them to re-pin them safely.

For example, if you want to re-pin all objects found by the 'lost-pins' mode,
you can do:

```bash
$ ipfs-see-all lost-pins > pins-out
$ ipfs pin add < pins-out
```

Do note that if you have lost pins and ran `ipfs repo gc`, the data referenced
by those pins may no longer be available.

### content-stat
The 'content-stat' mode iterates through every object in your repo and searches
for roots of graphs. This means any node that is not linked to by any other
node. Once it has this set, it attempts to classify each of them. Checking
if each object is a unixfs directory, file or other such type. It then
prints out information on each of those root objects.

This mode can be used to get an idea of what sort of content is in your repo.

## Installing

Prebuilt binaries are available at https://dist.ipfs.io/#ipfs-see-all

### From Source
Clone it down anywhere and run `make`.

```
$ git clone https://github.com/whyrusleeping/ipfs-see-all
$ cd ipfs-see-all
$ make
```

### License
MIT
