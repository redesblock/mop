package pipeline

import (
	"encoding/binary"
	"errors"

	"github.com/redesblock/hop/core/swarm"
)

var errInconsistentRefs = errors.New("inconsistent reference lengths in level")

type hashTrieWriter struct {
	branching  int
	chunkSize  int
	refSize    int
	fullChunk  int    // full chunk size in terms of the data represented in the buffer (span+refsize)
	cursors    []int  // level cursors, key is level. level 0 is data level
	buffer     []byte // keeps all level data
	pipelineFn pipelineFunc
}

func newHashTrieWriter(chunkSize, branching, refLen int, pipelineFn pipelineFunc) chainWriter {
	return &hashTrieWriter{
		cursors:    make([]int, 9),
		buffer:     make([]byte, swarm.ChunkWithSpanSize*9*2), // double size as temp workaround for weak calculation of needed buffer space
		branching:  branching,
		chunkSize:  chunkSize,
		refSize:    refLen,
		fullChunk:  (refLen + swarm.SpanSize) * branching,
		pipelineFn: pipelineFn,
	}
}

// accepts writes of hashes from the previous writer in the chain, by definition these writes
// are on level 1
func (h *hashTrieWriter) chainWrite(p *pipeWriteArgs) error {
	return h.writeToLevel(1, p.span, p.ref)
}

func (h *hashTrieWriter) writeToLevel(level int, span, ref []byte) error {
	copy(h.buffer[h.cursors[level]:h.cursors[level]+len(span)], span) //copy the span slongside
	h.cursors[level] += len(span)
	copy(h.buffer[h.cursors[level]:h.cursors[level]+len(ref)], ref)
	h.cursors[level] += len(ref)

	howLong := (h.refSize + swarm.SpanSize) * h.branching
	if h.levelSize(level) == howLong {
		return h.wrapFullLevel(level)
	}
	return nil
}

// wrapLevel wraps an existing level and writes the resulting hash to the following level
// then truncates the current level data by shifting the cursors.
// Steps are performed in the following order:
//	 - take all of the data in the current level
//	 - break down span and hash data
//	 - sum the span size, concatenate the hash to the buffer
//	 - call the short pipeline with the span and the buffer
//	 - get the hash that was created, append it one level above, and if necessary, wrap that level too
//	 - remove already hashed data from buffer

// assumes that the function has been called when refsize+span*branching has been reached
func (h *hashTrieWriter) wrapFullLevel(level int) error {
	data := h.buffer[h.cursors[level+1]:h.cursors[level]]
	sp := uint64(0)
	var hashes []byte
	for i := 0; i < len(data); i += h.refSize + 8 {
		// sum up the spans of the level, then we need to bmt them and store it as a chunk
		// then write the chunk address to the next level up
		sp += binary.LittleEndian.Uint64(data[i : i+8])
		hash := data[i+8 : i+h.refSize+8]
		hashes = append(hashes, hash...)
	}
	spb := make([]byte, 8)
	binary.LittleEndian.PutUint64(spb, sp)
	hashes = append(spb, hashes...)
	writer := h.pipelineFn()
	args := pipeWriteArgs{
		data: hashes,
		span: spb,
	}
	err := writer.chainWrite(&args)
	if err != nil {
		return err
	}
	err = h.writeToLevel(level+1, args.span, args.ref)
	if err != nil {
		return err
	}

	// this "truncates" the current level that was wrapped
	// by setting the cursors to the cursors of one level above
	h.cursors[level] = h.cursors[level+1]
	return nil
}

// pulls and potentially wraps all levels up to target
func (h *hashTrieWriter) hoistLevels(target int) ([]byte, error) {
	oneRef := h.refSize + swarm.SpanSize
	for i := 1; i < target; i++ {
		l := h.levelSize(i)
		if l%oneRef != 0 {
			return nil, errInconsistentRefs
		}
		switch {
		case l == 0:
			continue
		case l == h.fullChunk:
			err := h.wrapFullLevel(i)
			if err != nil {
				return nil, err
			}
		case l == oneRef:
			h.cursors[i+1] = h.cursors[i]
		default:
			// more than 0 but smaller than chunk size - wrap the level to the one above it
			err := h.wrapFullLevel(i)
			if err != nil {
				return nil, err
			}
		}
	}
	level := target
	tlen := h.levelSize(target)
	data := h.buffer[h.cursors[level+1]:h.cursors[level]]
	if tlen%oneRef != 0 {
		return nil, errInconsistentRefs
	}
	if tlen == oneRef {
		return data[8:], nil
	}

	// here we are still with possible length of more than one ref in the highest+1 level
	sp := uint64(0)
	var hashes []byte
	for i := 0; i < len(data); i += h.refSize + 8 {
		// sum up the spans of the level, then we need to bmt them and store it as a chunk
		// then write the chunk address to the next level up
		sp += binary.LittleEndian.Uint64(data[i : i+8])
		hash := data[i+8 : i+h.refSize+8]
		hashes = append(hashes, hash...)
	}
	spb := make([]byte, 8)
	binary.LittleEndian.PutUint64(spb, sp)
	hashes = append(spb, hashes...)
	writer := h.pipelineFn()
	args := pipeWriteArgs{
		data: hashes,
		span: spb,
	}
	err := writer.chainWrite(&args)

	return args.ref, err
}

func (h *hashTrieWriter) levelSize(level int) int {
	if level == 8 {
		return h.cursors[level]
	}
	return h.cursors[level] - h.cursors[level+1]
}

func (h *hashTrieWriter) sum() ([]byte, error) {
	// look from the top down, to look for the highest hash of a balanced tree
	// then, whatever is in the levels below that is necessarily unbalanced,
	// so, we'd like to reduce those levels to one hash, then wrap it together
	// with the balanced tree hash, to produce the root chunk
	highest := 1
	for i := 8; i > 0; i-- {
		if h.levelSize(i) > 0 && i > highest {
			highest = i
		}
	}
	return h.hoistLevels(highest)
}
