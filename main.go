package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	blocks "github.com/ipfs/go-ipfs/blocks/blockstore"
	bserv "github.com/ipfs/go-ipfs/blockservice"
	"github.com/ipfs/go-ipfs/merkledag"
	"github.com/ipfs/go-ipfs/pin"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	ft "github.com/ipfs/go-ipfs/unixfs"
	cid "gx/ipfs/QmakyCk6Vnn16WEKjbkxieZmM2YLTzkFWizbmGowoYPjro/go-cid"
)

type objectInfo struct {
	Cid       *cid.Cid
	Type      string
	TotalSize uint64
	Pinned    bool
}

type objectInfos []objectInfo

func (ois objectInfos) Len() int {
	return len(ois)
}

func (ois objectInfos) Swap(i, j int) {
	ois[i], ois[j] = ois[j], ois[i]
}

func (ois objectInfos) Less(i, j int) bool {
	if ois[i].Type == "unknown" {
		if ois[j].Type != "unknown" {
			return false
		}

		if ois[i].Pinned && !ois[j].Pinned {
			return true
		}

		return ois[i].TotalSize > ois[j].TotalSize
	}

	if ois[j].Type == "unknown" {
		return true
	}

	if ois[i].Pinned && !ois[j].Pinned {
		return true
	}

	return ois[i].TotalSize > ois[j].TotalSize
}

func fatal(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

func main() {
	p, err := fsrepo.BestKnownPath()
	if err != nil {
		fatal(err)
	}

	r, err := fsrepo.Open(p)
	if err != nil {
		fatal(err)
	}

	bs := blocks.NewBlockstore(r.Datastore())
	keys, err := bs.AllKeysChan(context.Background())
	if err != nil {
		fatal(err)
	}

	dag := merkledag.NewDAGService(bserv.New(bs, nil))

	pinner, err := pin.LoadPinner(r.Datastore(), dag, dag)
	if err != nil {
		fatal(err)
	}

	recpins := cid.NewSet()
	for _, c := range pinner.RecursiveKeys() {
		recpins.Add(c)
	}

	fmt.Printf("%s: started processing keys...\n", time.Now())
	allKeys := cid.NewSet()
	for bk := range keys {
		c := cid.NewCidV0(bk.ToMultihash())
		allKeys.Add(c)
	}

	fmt.Printf("%s: initial key gathering complete, now finding graph roots.\n", time.Now())
	for _, c := range allKeys.Keys() {
		nd, err := dag.Get(context.Background(), c)
		if err != nil {
			fmt.Printf("error reading dag node (%s): %s\n", c, err)
			continue
		}

		for _, lnk := range nd.Links {
			c := cid.NewCidV0(lnk.Hash)
			if !recpins.Has(c) {
				allKeys.Remove(cid.NewCidV0(lnk.Hash))
			}
		}
	}

	fmt.Printf("%s: root selection complete, classifying resulting objects\n", time.Now())
	var output []objectInfo
	// just left with roots now
	for _, c := range allKeys.Keys() {
		nd, err := dag.Get(context.Background(), c)
		if err != nil {
			fmt.Printf("error reading dag node (%s): %s\n", c, err)
			continue
		}

		size, err := nd.Size()
		if err != nil {
			fmt.Println("error getting size of object: ", err)
		}

		oi := objectInfo{
			Cid:       c,
			Pinned:    recpins.Has(c),
			TotalSize: size,
		}

		fsn, err := ft.FSNodeFromBytes(nd.Data())
		if err == nil {
			oi.Type = "unixfs-" + fsn.Type.String()
		} else {
			oi.Type = "unknown"
		}

		output = append(output, oi)
	}

	fmt.Printf("%s: classification complete, sorting output...\n", time.Now())
	sort.Sort(objectInfos(output))

	outputObjectInfos(dag, output)
}

func outputObjectInfos(dag merkledag.DAGService, ois []objectInfo) {
	fmt.Println("Hash                  Type\tSize\tPinned(recursively)")
	for _, oi := range ois {
		fmt.Printf("%s %s\t%d\t%t\n", oi.Cid, oi.Type, oi.TotalSize, oi.Pinned)
		if oi.Type == "unixfs-Directory" {
			nd, err := dag.Get(context.Background(), oi.Cid)
			if err != nil {
				fmt.Println("Error fetching node: ", err)
				continue
			}

			nshow := 5
			if len(nd.Links) < 5 {
				nshow = len(nd.Links)
			}

			fmt.Print("\tDirents: [ ")
			for _, lnk := range nd.Links[:nshow] {
				fmt.Printf("%q ", lnk.Name)
			}

			fmt.Print("]")

			if len(nd.Links) > 5 {
				fmt.Print(" ...")
			}
			fmt.Println()
		}
	}
}
