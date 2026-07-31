package id

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

type ID int64

const (
	nodeBits = 10
	seqBits  = 12

	nodeMask = (1 << nodeBits) - 1
	seqMask  = (1 << seqBits) - 1
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

var (
	mu       sync.Mutex
	nodeID   int64
	lastTime int64
	sequence int64
)

func init() {
	node, err := strconv.ParseInt(os.Getenv("SNOWFLAKE_NODE_ID"), 10, 64)
	if err != nil || node < 0 || node > nodeMask {
		node = 0
	}
	nodeID = node
}

func New() ID {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UTC().Sub(epoch).Milliseconds()
	for now < lastTime {
		time.Sleep(time.Duration(lastTime-now) * time.Millisecond)
		now = time.Now().UTC().Sub(epoch).Milliseconds()
	}
	if now == lastTime {
		sequence = (sequence + 1) & seqMask
		if sequence == 0 {
			for now <= lastTime {
				time.Sleep(time.Millisecond)
				now = time.Now().UTC().Sub(epoch).Milliseconds()
			}
			sequence = 0
		}
	} else {
		sequence = 0
	}
	lastTime = now

	return ID((now << (nodeBits + seqBits)) | (nodeID << seqBits) | sequence)
}

func Parse(s string) (ID, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", s)
	}
	return ID(n), nil
}

func (i ID) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *ID) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		n, err := Parse(s)
		if err != nil {
			return err
		}
		*i = n
		return nil
	}
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("invalid id: %s", string(b))
	}
	*i = ID(n)
	return nil
}
